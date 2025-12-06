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
	rssGenerator  *rss.Generator
	inititialized sync.Once
}

var Posts = &PostsCache{}

func (c *PostsCache) load(renderedPosts []renderer.RenderedPost) {
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

	if c.rssGenerator != nil {
		c.rssGenerator.Generate(c.allPosts)
	}
}

func (c *PostsCache) SetRSSGenerator(gen *rss.Generator) {
	c.rssGenerator = gen
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

func (c *PostsCache) Initialize(fsys embed.FS, rssGen *rss.Generator, ctx context.Context) {
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

		c.rssGenerator = rssGen
		c.load(cachedPosts)
	})
}
