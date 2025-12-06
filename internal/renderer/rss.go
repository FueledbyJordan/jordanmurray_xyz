package renderer

import (
	"fmt"
	"jordanmurray.xyz/site/internal/models"
	"jordanmurray.xyz/site/internal/utils"
)

type RenderedRSSFeed struct {
	models.RSSFeed
	CompressedFeed []byte
}

func NewRenderedRSSFeed(rssFeed models.RSSFeed) (RenderedRSSFeed, error) {
	compressedFeed, err := utils.Compress(rssFeed.Feed, utils.DefaultCompression)
	if err != nil {
		return RenderedRSSFeed{}, fmt.Errorf("error compressing rss feed: %w", err)
	}

	return RenderedRSSFeed{
		RSSFeed:        rssFeed,
		CompressedFeed: compressedFeed,
	}, nil
}

func (r RenderedRSSFeed) Data() []byte {
	return r.Feed
}

func (r RenderedRSSFeed) CompressedData() []byte {
	return r.CompressedFeed
}

func (r RenderedRSSFeed) ContentType() string {
	return "application/rss+xml; charset=utf-8"
}
