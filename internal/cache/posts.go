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
	"jordanmurray.xyz/site/internal/rss"
)

type PostsCache struct {
	allPosts      []models.Post
	postBySlug    map[string]renderer.RenderedPost
	rss           rss.RenderedRSS
	inititialized sync.Once
}

var Posts = &PostsCache{}

func (c *PostsCache) load(renderedPosts []renderer.RenderedPost, rssConfig rss.Config) error {
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

	renderedRSS, err := rss.Generate(c.allPosts, rssConfig)
	if err != nil {
		return fmt.Errorf("failed to generate rss: %w", err)
	}
	c.rss = renderedRSS

	return nil
}

func (c *PostsCache) GetAllPosts() []models.Post {
	return c.allPosts
}

func (c *PostsCache) GetPostBySlug(slug string) (renderer.RenderedPost, error) {
	post, ok := c.postBySlug[slug]
	if !ok {
		return renderer.RenderedPost{}, errors.New("could not find post")
	}

	return post, nil
}

func (c *PostsCache) GetRSS() rss.RenderedRSS {
	return c.rss
}

func (c *PostsCache) Initialize(fsys embed.FS, rssConfig rss.Config, ctx context.Context) {
	c.inititialized.Do(func() {
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

		if err := c.load(cachedPosts, rssConfig); err != nil {
			panic(fmt.Errorf("error loading cache: %w", err))
		}
	})
}
