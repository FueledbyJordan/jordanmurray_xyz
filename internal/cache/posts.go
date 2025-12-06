package cache

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"jordanmurray.xyz/site/internal/models"
	"jordanmurray.xyz/site/internal/rss"
	"jordanmurray.xyz/site/internal/utils"
)

type CachedPost struct {
	models.Post
	RenderedHTML       []byte
	RenderedHTMLBrotli []byte
}

type Renderer interface {
	Render(post *models.Post) ([]byte, error)
}

type PostsCache struct {
	allPosts     []models.Post
	postBySlug   map[string]*CachedPost
	once         sync.Once
	contentFS    fs.FS
	renderer     Renderer
	rssGenerator *rss.Generator
}

var Posts = &PostsCache{}

func (c *PostsCache) Init() {
	c.once.Do(func() {
		if c.contentFS == nil {
			panic("content filesystem not set")
		}

		posts := []models.Post{}
		slugMap := make(map[string]*CachedPost)

		entries, err := fs.ReadDir(c.contentFS, "content/reflections")
		if err != nil {
			panic(fmt.Errorf("error reading reflections directory: %w", err))
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			path := filepath.Join("content/reflections", entry.Name())
			post, err := models.LoadPostFromFS(c.contentFS, path)
			if err != nil {
				panic(fmt.Errorf("error loading post from %s: %w", path, err))
			}

			posts = append(posts, *post)

			cachedPost := &CachedPost{Post: *post}
			if c.renderer != nil {
				rendered, err := c.renderer.Render(post)
				if err != nil {
					panic(fmt.Errorf("error rendering post %s: %w", post.Slug, err))
				}

				cachedPost.RenderedHTML = rendered

				compressed, err := utils.Compress(rendered, utils.DefaultCompression)
				if err != nil {
					panic(fmt.Errorf("Error compressing post %s: %v", post.Slug, err))
				}

				cachedPost.RenderedHTMLBrotli = compressed
			}

			slugMap[post.Slug] = cachedPost
		}

		slices.SortFunc(posts, func(a, b models.Post) int {
			return b.PublishedAt.Compare(a.PublishedAt)
		})

		c.allPosts = posts
		c.postBySlug = slugMap
		c.rssGenerator.Generate(c.allPosts)
	})
}

func (c *PostsCache) SetContentFS(fsys fs.FS) {
	c.contentFS = fsys
}

func (c *PostsCache) SetRenderer(r Renderer) {
	c.renderer = r
}

func (c *PostsCache) SetRSSGenerator(gen *rss.Generator) {
	c.rssGenerator = gen
}

func (c *PostsCache) GetAllPosts() []models.Post {
	c.Init()
	return c.allPosts
}

func (c *PostsCache) GetPostBySlug(slug string) *CachedPost {
	c.Init()
	return c.postBySlug[slug]
}
