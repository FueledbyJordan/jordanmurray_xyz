package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"jordanmurray.xyz/site/internal/models"
	"jordanmurray.xyz/site/internal/utils"
)

type rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []item `xml:"item"`
}

type item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

type Config struct {
	BaseURL     string
	Title       string
	Description string
}

type Generator struct {
	Config
	rssFeed           []byte
	compressedRssFeed []byte
}

func New(cfg Config) *Generator {
	return &Generator{
		Config: cfg,
	}
}

func (g *Generator) Generate(posts []models.Post) error {
	var items []item
	var lastBuildDate time.Time

	for _, post := range posts {
		postPath := strings.Join([]string{g.BaseURL, post.Slug}, "/")
		items = append(items, item{
			Title:       post.Title,
			Link:        postPath,
			Description: post.Excerpt,
			PubDate:     post.PublishedAt.Format(time.RFC1123Z),
			GUID:        postPath,
		})

		if post.PublishedAt.After(lastBuildDate) {
			lastBuildDate = post.PublishedAt
		}
	}

	if lastBuildDate.IsZero() && len(posts) > 0 {
		lastBuildDate = time.Now()
	}

	feed := rss{
		Version: "2.0",
		Channel: channel{
			Title:         g.Title,
			Link:          g.BaseURL,
			Description:   g.Description,
			Language:      "en-us",
			LastBuildDate: lastBuildDate.Format(time.RFC1123Z),
			Items:         items,
		},
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	if err := encoder.Encode(feed); err != nil {
		return fmt.Errorf("failed to encode rss feed: %w", err)
	}

	rssFeed := buf.Bytes()
	g.rssFeed = rssFeed

	compressed, err := utils.Compress(buf.Bytes(), utils.DefaultCompression)
	if err != nil {
		return fmt.Errorf("failed to compress rss feed: %w", err)
	}
	g.compressedRssFeed = compressed

	return nil
}

func (g *Generator) RssFeed() []byte {
	return g.rssFeed
}

func (g *Generator) CompressedRssFeed() []byte {
	return g.compressedRssFeed
}
