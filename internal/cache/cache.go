package cache

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"sync"

	"jordanmurray.xyz/site/internal/models"
	"jordanmurray.xyz/site/internal/renderer"
)

type Cache struct {
	allPosts      []models.Post
	postBySlug    map[string]renderer.RenderedPost
	rss           renderer.RenderedRSSFeed
	inititialized sync.Once
}

var cache = &Cache{}

func Get() (*Cache, error) {
	if cache == nil {
		return nil, errors.New("cache is nil")
	}
	return cache, nil
}

func (c *Cache) load(renderedPosts []renderer.RenderedPost, rssConfig models.RSSConfig) error {
	posts := make([]models.Post, len(renderedPosts))
	slugMap := make(map[string]renderer.RenderedPost)

	for i, cp := range renderedPosts {
		posts[i] = cp.Post
		slugMap[cp.Slug] = cp
	}

	slices.SortFunc(posts, func(a, b models.Post) int {
		return b.PublishedAt.Compare(a.PublishedAt)
	})

	c.allPosts = posts
	c.postBySlug = slugMap

	rssFeed := models.NewRSSFeed(rssConfig)
	err := rssFeed.FromPosts(c.allPosts)
	if err != nil {
		return fmt.Errorf("failed to generate rss: %w", err)
	}

	renderedRssFeed, err := renderer.NewRenderedRSSFeed(rssFeed)
	if err != nil {
		return fmt.Errorf("failed to compress rss: %w", err)
	}
	c.rss = renderedRssFeed

	return nil
}

func (c *Cache) AllPosts() []models.Post {
	return c.allPosts
}

func (c *Cache) PostBySlug(slug string) (renderer.RenderedPost, error) {
	post, ok := c.postBySlug[slug]
	if !ok {
		return renderer.RenderedPost{}, errors.New("could not find post")
	}

	return post, nil
}

func (c *Cache) RSS() renderer.RenderedRSSFeed {
	return c.rss
}

func Initialize(fsys embed.FS, rssConfig models.RSSConfig, ctx context.Context) {
	cache.inititialized.Do(func() {
		entries, err := fs.ReadDir(fsys, "content/reflections")
		if err != nil {
			panic(fmt.Errorf("error reading reflections directory: %w", err))
		}

		var cachedPosts []renderer.RenderedPost
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
				continue
			}

			path := filepath.Join("content/reflections", entry.Name())
			post, err := models.LoadPostFromFS(fsys, path)
			if err != nil {
				panic(fmt.Errorf("error loading post from %s: %w", path, err))
			}

			cachedPost, err := renderer.NewRenderedPost(post, ctx)
			if err != nil {
				panic(fmt.Errorf("error caching post %s: %w", post.Slug, err))
			}

			cachedPosts = append(cachedPosts, cachedPost)
		}

		if err := cache.load(cachedPosts, rssConfig); err != nil {
			panic(fmt.Errorf("error loading cache: %w", err))
		}
	})
}
