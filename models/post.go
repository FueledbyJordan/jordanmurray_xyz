package models

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

type Post struct {
	ID          string
	Title       string
	Slug        string
	Author      string
	PublishedAt time.Time
	Content     string
	Excerpt     string
	Tags        []string
}

type FrontMatter struct {
	Title       string    `yaml:"title"`
	Author      string    `yaml:"author"`
	PublishedAt time.Time `yaml:"published_at"`
	Excerpt     string    `yaml:"excerpt"`
	Tags        []string  `yaml:"tags"`
}

type postsCache struct {
	allPosts   []Post
	postBySlug map[string]*Post
	once       sync.Once
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

func loadPostFromFile(path string) (*Post, error) {
	content, err := os.ReadFile(path)
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

func (cache *postsCache) init() {
	cache.once.Do(func() {
		posts := []Post{}
		slugMap := make(map[string]*Post)

		reflectionsDir := "content/reflections"
		files, err := filepath.Glob(filepath.Join(reflectionsDir, "*.md"))
		if err != nil {
			log.Printf("Error reading reflections directory: %v", err)
			return
		}

		for _, file := range files {
			post, err := loadPostFromFile(file)
			if err != nil {
				log.Printf("Error loading post from %s: %v", file, err)
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
	})
}

func GetAllPosts() []Post {
	cache.init()
	return cache.allPosts
}

func GetPostBySlug(slug string) *Post {
	cache.init()
	return cache.postBySlug[slug]
}
