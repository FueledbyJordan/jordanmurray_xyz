package models

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
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
	ID      string
	Slug    string
	Content string
	FrontMatter
}

type FrontMatter struct {
	Title       string    `yaml:"title"`
	Author      string    `yaml:"author"`
	PublishedAt time.Time `yaml:"published_at"`
	Excerpt     string    `yaml:"excerpt"`
	Tags        []string  `yaml:"tags"`
}

func parseFrontMatter(content []byte) (FrontMatter, []byte, error) {
	var fm FrontMatter

	if !bytes.HasPrefix(content, []byte("---\n")) {
		return fm, content, fmt.Errorf("no front matter found")
	}

	parts := bytes.SplitN(content[4:], []byte("\n---\n"), 2)
	if len(parts) != 2 {
		return fm, content, fmt.Errorf("invalid front matter format")
	}

	if err := yaml.Unmarshal(parts[0], &fm); err != nil {
		return fm, []byte{}, fmt.Errorf("failed to parse front matter: %w", err)
	}

	return fm, parts[1], nil
}

func renderMarkdown(markdown []byte) ([]byte, error) {
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
	if err := md.Convert(markdown, &buf); err != nil {
		return []byte{}, fmt.Errorf("failed to render markdown: %w", err)
	}

	return buf.Bytes(), nil
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
		ID: slug,
		FrontMatter: FrontMatter{
			Title:       fm.Title,
			Author:      fm.Author,
			PublishedAt: fm.PublishedAt,
			Excerpt:     fm.Excerpt,
			Tags:        fm.Tags,
		},
		Slug:    slug,
		Content: string(htmlContent),
	}, nil
}
