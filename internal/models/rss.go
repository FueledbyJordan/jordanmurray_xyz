package models

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
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

type RSSConfig struct {
	BaseURL     string
	Title       string
	Description string
}

type RSSFeed struct {
	RSSConfig
	Feed []byte
}

func NewRSSFeed(cfg RSSConfig) RSSFeed {
	return RSSFeed{
		RSSConfig: cfg,
	}
}

func (r *RSSFeed) FromPosts(posts []Post) error {
	var items []item
	var lastBuildDate time.Time

	for _, post := range posts {
		postPath := strings.Join([]string{r.BaseURL, post.Slug}, "/")
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
			Title:         r.Title,
			Link:          r.BaseURL,
			Description:   r.Description,
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

	r.Feed = buf.Bytes()

	return nil
}

func (r RSSFeed) Empty() bool {
	return len(r.Feed) == 0
}
