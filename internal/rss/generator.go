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

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

type Generator struct {
	baseURL        string
	feed           []byte
	feedBrotli     []byte
	title          string
	description    string
}

func NewGenerator(URL, title, description string) *Generator {
	return &Generator{
		baseURL:     URL,
		title:       title,
		description: description,
	}
}

func (g *Generator) Generate(posts []models.Post) error {
	var items []Item
	var lastBuildDate time.Time

	for _, post := range posts {
		postPath := strings.Join([]string{g.baseURL, post.Slug}, "/")
		items = append(items, Item{
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

	feed := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:         g.title,
			Link:          g.baseURL,
			Description:   g.description,
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

	g.feed = buf.Bytes()

	// Also create compressed version
	compressed, err := utils.Compress(buf.Bytes(), utils.DefaultCompression)
	if err != nil {
		return fmt.Errorf("failed to compress rss feed: %w", err)
	}
	g.feedBrotli = compressed

	return nil
}

func (g *Generator) GetFeed() []byte {
	return g.feed
}

func (g *Generator) GetFeedBrotli() []byte {
	return g.feedBrotli
}
