package models

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
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

func LoadPostFromFS(fsys fs.FS, path string) (*Post, error) {
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
