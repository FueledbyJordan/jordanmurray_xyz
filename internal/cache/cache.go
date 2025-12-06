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

var cache = &Cache{}

type Cache struct {
	allPosts   []models.Post
	postBySlug map[string]renderer.RenderedPost
	rss        renderer.RenderedRSSFeed
	once       sync.Once
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

func (c *Cache) storePosts(renderedPosts []renderer.RenderedPost) {
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
}

func (c *Cache) constructRss(rssConfig models.RSSConfig) error {
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

func Hydrate(fsys embed.FS, rssConfig models.RSSConfig, ctx context.Context) {
	cache.once.Do(func() {
		cachedPosts, err := loadRenderedPosts(fsys, ctx)
		if err != nil {
			panic(fmt.Errorf("error loading posts: %w", err))
		}

		cache.storePosts(cachedPosts)
		if err := cache.constructRss(rssConfig); err != nil {
			panic(fmt.Errorf("error caching rss: %w", err))
		}
	})
}

func Get() (*Cache, error) {
	if cache == nil {
		return nil, errors.New("cache is nil")
	}
	return cache, nil
}

func loadRenderedPosts(fsys embed.FS, ctx context.Context) ([]renderer.RenderedPost, error) {
	var cachedPosts []renderer.RenderedPost

	err := fs.WalkDir(fsys, "content/reflections", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking error: %w", err)
		}

		if d.IsDir() || filepath.Ext(d.Name()) != ".md" {
			return nil
		}

		cachedPost, err := loadAndRenderPost(fsys, path, ctx)
		if err != nil {
			return fmt.Errorf("rendering posts: %w", err)
		}

		cachedPosts = append(cachedPosts, cachedPost)
		return nil
	})

	return cachedPosts, err
}

func loadAndRenderPost(fsys embed.FS, path string, ctx context.Context) (renderer.RenderedPost, error) {
	post, err := models.LoadPostFromFS(fsys, path)
	if err != nil {
		return renderer.RenderedPost{}, fmt.Errorf("error loading post from %s: %w", path, err)
	}

	cachedPost, err := renderer.NewRenderedPost(post, ctx)
	if err != nil {
		return renderer.RenderedPost{}, fmt.Errorf("error rendering post %s: %w", post.Slug, err)
	}

	return cachedPost, nil
}
