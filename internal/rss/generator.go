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

type RenderedRSS struct {
	Feed           []byte
	CompressedFeed []byte
}

func Generate(posts []models.Post, cfg Config) (RenderedRSS, error) {
	var items []item
	var lastBuildDate time.Time

	for _, post := range posts {
		postPath := strings.Join([]string{cfg.BaseURL, post.Slug}, "/")
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
			Title:         cfg.Title,
			Link:          cfg.BaseURL,
			Description:   cfg.Description,
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
		return RenderedRSS{}, fmt.Errorf("failed to encode rss feed: %w", err)
	}

	rssFeed := buf.Bytes()

	compressed, err := utils.Compress(rssFeed, utils.DefaultCompression)
	if err != nil {
		return RenderedRSS{}, fmt.Errorf("failed to compress rss feed: %w", err)
	}

	return RenderedRSS{
		Feed:           rssFeed,
		CompressedFeed: compressed,
	}, nil
}

func (r RenderedRSS) Empty() bool {
	return len(r.Feed) == 0
}

func (r RenderedRSS) Data() []byte {
	return r.Feed
}

func (r RenderedRSS) CompressedData() []byte {
	return r.CompressedFeed
}

func (r RenderedRSS) ContentType() string {
	return "application/rss+xml; charset=utf-8"
}
