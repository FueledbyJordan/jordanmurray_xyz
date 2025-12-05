package rss

import (
	"bytes"
	"encoding/xml"
	"log"
	"time"

	"github.com/andybalholm/brotli"
	"jordanmurray.xyz/site/internal/models"
)

// RSS feed types
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
	baseURL       string
	feed          []byte
	feedBrotli    []byte
	title         string
	description   string
}

func NewGenerator(baseURL, title, description string) *Generator {
	return &Generator{
		baseURL:     baseURL,
		title:       title,
		description: description,
	}
}

func (g *Generator) Generate(posts []models.Post) {
	var items []Item
	var lastBuildDate time.Time

	for _, post := range posts {
		items = append(items, Item{
			Title:       post.Title,
			Link:        g.baseURL + "/reflections/" + post.Slug,
			Description: post.Excerpt,
			PubDate:     post.PublishedAt.Format(time.RFC1123Z),
			GUID:        g.baseURL + "/reflections/" + post.Slug,
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
		log.Printf("Error encoding RSS feed: %v", err)
		return
	}

	g.feed = buf.Bytes()

	// Compress with brotli
	var compressed bytes.Buffer
	writer := brotli.NewWriterLevel(&compressed, 6)
	if _, err := writer.Write(g.feed); err != nil {
		log.Printf("Error compressing RSS feed: %v", err)
		return
	}
	if err := writer.Close(); err != nil {
		log.Printf("Error closing RSS compressor: %v", err)
		return
	}

	g.feedBrotli = compressed.Bytes()

	log.Printf("Pre-generated RSS feed: %d bytes (uncompressed) -> %d bytes (brotli) [%.1f%% reduction]",
		len(g.feed),
		len(g.feedBrotli),
		100.0*(1.0-float64(len(g.feedBrotli))/float64(len(g.feed))))
}

func (g *Generator) GetFeed() []byte {
	return g.feed
}

func (g *Generator) GetFeedBrotli() []byte {
	return g.feedBrotli
}
