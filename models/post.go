package models

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
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
	Content     string // HTML content rendered from markdown
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

// parseFrontMatter separates YAML front matter from markdown content
func parseFrontMatter(content []byte) (FrontMatter, string, error) {
	var fm FrontMatter

	// Check if content starts with ---
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return fm, string(content), fmt.Errorf("no front matter found")
	}

	// Find the end of front matter
	parts := bytes.SplitN(content[4:], []byte("\n---\n"), 2)
	if len(parts) != 2 {
		return fm, string(content), fmt.Errorf("invalid front matter format")
	}

	// Parse YAML front matter
	if err := yaml.Unmarshal(parts[0], &fm); err != nil {
		return fm, "", fmt.Errorf("failed to parse front matter: %w", err)
	}

	return fm, string(parts[1]), nil
}

// renderMarkdown converts markdown to HTML with syntax highlighting
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
			goldmarkhtml.WithUnsafe(), // Allow raw HTML in markdown
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return buf.String(), nil
}

// loadPostFromFile loads a post from a markdown file
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

	// Generate slug from filename
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

// GetAllPosts loads all posts from the content/reflections directory
func GetAllPosts() []Post {
	posts := []Post{}

	reflectionsDir := "content/reflections"
	files, err := filepath.Glob(filepath.Join(reflectionsDir, "*.md"))
	if err != nil {
		log.Printf("Error reading reflections directory: %v", err)
		return posts
	}

	for _, file := range files {
		post, err := loadPostFromFile(file)
		if err != nil {
			log.Printf("Error loading post from %s: %v", file, err)
			continue
		}
		posts = append(posts, *post)
	}

	// Sort posts by published date (newest first)
	for i := 0; i < len(posts); i++ {
		for j := i + 1; j < len(posts); j++ {
			if posts[i].PublishedAt.Before(posts[j].PublishedAt) {
				posts[i], posts[j] = posts[j], posts[i]
			}
		}
	}

	return posts
}

func GetPostBySlug(slug string) *Post {
	posts := GetAllPosts()
	for _, post := range posts {
		if post.Slug == slug {
			return &post
		}
	}
	return nil
}
