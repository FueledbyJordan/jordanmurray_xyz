package cache

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"

	"jordanmurray.xyz/site/internal/models"
	"jordanmurray.xyz/site/internal/renderer"
	"jordanmurray.xyz/site/internal/rss"
	"jordanmurray.xyz/site/internal/utils"
)

type CachedPost struct {
	models.Post
	RenderedHTML       []byte
	RenderedHTMLBrotli []byte
}

type PostsCache struct {
	allPosts     []models.Post
	postBySlug   map[string]*CachedPost
	rssGenerator *rss.Generator
}

var Posts = &PostsCache{}

func NewCachedPost(post models.Post, ctx context.Context) (*CachedPost, error) {
	cachedPost := &CachedPost{Post: post}
	rendered, err := renderer.PostRenderer{Post: post}.Render(ctx)
	if err != nil {
		return nil, fmt.Errorf("error rendering post: %w", err)
	}

	cachedPost.RenderedHTML = rendered

	compressed, err := utils.Compress(rendered, utils.DefaultCompression)
	if err != nil {
		return nil, fmt.Errorf("error compressing post: %w", err)
	}

	cachedPost.RenderedHTMLBrotli = compressed
	return cachedPost, nil
}

func (c *PostsCache) Load(cachedPosts []*CachedPost) {
	posts := make([]models.Post, len(cachedPosts))
	slugMap := make(map[string]*CachedPost)

	for i, cp := range cachedPosts {
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

func (c *PostsCache) GetPostBySlug(slug string) *CachedPost {
	return c.postBySlug[slug]
}

func LoadPosts(fsys embed.FS, ctx context.Context) ([]*CachedPost, error) {
	entries, err := fs.ReadDir(fsys, "content/reflections")
	if err != nil {
		return nil, fmt.Errorf("error reading reflections directory: %w", err)
	}

	var cachedPosts []*CachedPost
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		path := filepath.Join("content/reflections", entry.Name())
		post, err := models.LoadPostFromFS(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("error loading post from %s: %w", path, err)
		}

		cachedPost, err := NewCachedPost(*post, ctx)
		if err != nil {
			return nil, fmt.Errorf("error caching post %s: %w", post.Slug, err)
		}

		cachedPosts = append(cachedPosts, cachedPost)
	}

	return cachedPosts, nil
}
