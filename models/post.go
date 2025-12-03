package models

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/andybalholm/brotli"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

type Post struct {
	ID                 string
	Title              string
	Slug               string
	Author             string
	PublishedAt        time.Time
	Content            string
	Excerpt            string
	Tags               []string
	RenderedHTML       []byte // Pre-rendered complete HTML page
	RenderedHTMLBrotli []byte // Brotli compressed version
}

type FrontMatter struct {
	Title       string    `yaml:"title"`
	Author      string    `yaml:"author"`
	PublishedAt time.Time `yaml:"published_at"`
	Excerpt     string    `yaml:"excerpt"`
	Tags        []string  `yaml:"tags"`
}

type postsCache struct {
	allPosts      []Post
	postBySlug    map[string]*Post
	once          sync.Once
	renderFunc    func(*Post) ([]byte, error) // Injected function to render posts
	renderFuncSet bool
	contentFS     fs.FS // Embedded filesystem for content files
}

var cache = postsCache{}

func parseFrontMatter(content []byte) (FrontMatter, string, error) {
	var fm FrontMatter

	if !bytes.HasPrefix(content, []byte("---\n")) {
		return fm, string(content), fmt.Errorf("no front matter found")
	}

	parts := bytes.SplitN(content[4:], []byte("\n---\n"), 2)
	if len(parts) != 2 {
		return fm, string(content), fmt.Errorf("invalid front matter format")
	}

	if err := yaml.Unmarshal(parts[0], &fm); err != nil {
		return fm, "", fmt.Errorf("failed to parse front matter: %w", err)
	}

	return fm, string(parts[1]), nil
}

func renderMarkdown(markdown string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					html.WithClasses(true),
					html.TabWidth(2),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return buf.String(), nil
}

func loadPostFromFS(fsys fs.FS, path string) (*Post, error) {
	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fm, markdown, err := parseFrontMatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse front matter: %w", err)
	}

	htmlContent, err := renderMarkdown(markdown)
	if err != nil {
		return nil, fmt.Errorf("failed to render markdown: %w", err)
	}

	filename := filepath.Base(path)
	slug := strings.TrimSuffix(filename, filepath.Ext(filename))

	return &Post{
		ID:          slug,
		Title:       fm.Title,
		Slug:        slug,
		Author:      fm.Author,
		PublishedAt: fm.PublishedAt,
		Content:     htmlContent,
		Excerpt:     fm.Excerpt,
		Tags:        fm.Tags,
	}, nil
}

// SetRenderedHTML stores pre-rendered HTML and compresses it with brotli
func (p *Post) SetRenderedHTML(html []byte) error {
	p.RenderedHTML = html

	// Compress with brotli (quality 6 is a good balance of speed/compression)
	var compressed bytes.Buffer
	writer := brotli.NewWriterLevel(&compressed, 6)
	if _, err := writer.Write(html); err != nil {
		return fmt.Errorf("failed to compress: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close compressor: %w", err)
	}

	p.RenderedHTMLBrotli = compressed.Bytes()

	log.Printf("Pre-rendered post %s: %d bytes (uncompressed) -> %d bytes (brotli) [%.1f%% reduction]",
		p.Slug,
		len(p.RenderedHTML),
		len(p.RenderedHTMLBrotli),
		100.0*(1.0-float64(len(p.RenderedHTMLBrotli))/float64(len(p.RenderedHTML))))

	return nil
}

func (cache *postsCache) init() {
	cache.once.Do(func() {
		if cache.contentFS == nil {
			log.Printf("Error: content filesystem not set")
			return
		}

		posts := []Post{}
		slugMap := make(map[string]*Post)

		// Read all .md files from content/reflections
		entries, err := fs.ReadDir(cache.contentFS, "content/reflections")
		if err != nil {
			log.Printf("Error reading reflections directory: %v", err)
			return
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			path := filepath.Join("content/reflections", entry.Name())
			post, err := loadPostFromFS(cache.contentFS, path)
			if err != nil {
				log.Printf("Error loading post from %s: %v", path, err)
				continue
			}
			posts = append(posts, *post)
		}

		slices.SortFunc(posts, func(a, b Post) int {
			return b.PublishedAt.Compare(a.PublishedAt)
		})

		for i := range posts {
			slugMap[posts[i].Slug] = &posts[i]
		}

		cache.allPosts = posts
		cache.postBySlug = slugMap
		log.Printf("Loaded %d posts into cache", len(posts))

		// Pre-render posts if render function is set
		if cache.renderFuncSet && cache.renderFunc != nil {
			cache.preRenderAll()
		}
	})
}

func (cache *postsCache) preRenderAll() {
	for i := range cache.allPosts {
		post := &cache.allPosts[i]
		html, err := cache.renderFunc(post)
		if err != nil {
			log.Printf("Warning: failed to pre-render post %s: %v", post.Slug, err)
			continue
		}
		if err := post.SetRenderedHTML(html); err != nil {
			log.Printf("Warning: failed to compress post %s: %v", post.Slug, err)
		}
	}
}

// SetContentFS sets the embedded filesystem for reading content files
// Must be called before any posts are accessed
func SetContentFS(fsys fs.FS) {
	cache.contentFS = fsys
}

// SetRenderFunc sets the function used to render posts to HTML
// Must be called before any posts are accessed
func SetRenderFunc(fn func(*Post) ([]byte, error)) {
	cache.renderFunc = fn
	cache.renderFuncSet = true
}

func GetAllPosts() []Post {
	cache.init()
	return cache.allPosts
}

func GetPostBySlug(slug string) *Post {
	cache.init()
	return cache.postBySlug[slug]
}
