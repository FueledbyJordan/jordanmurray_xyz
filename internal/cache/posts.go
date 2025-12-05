package cache

import (
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"jordanmurray.xyz/site/internal/models"
	"jordanmurray.xyz/site/internal/rss"
)

type PostsCache struct {
	allPosts      []models.Post
	postBySlug    map[string]*models.Post
	once          sync.Once
	renderFunc    func(*models.Post) ([]byte, error)
	renderFuncSet bool
	contentFS     fs.FS
	rssGenerator  *rss.Generator
}

var Posts = &PostsCache{}

func (c *PostsCache) Init() {
	c.once.Do(func() {
		if c.contentFS == nil {
			log.Printf("Error: content filesystem not set")
			return
		}

		posts := []models.Post{}
		slugMap := make(map[string]*models.Post)

		// Read all .md files from content/reflections
		entries, err := fs.ReadDir(c.contentFS, "content/reflections")
		if err != nil {
			log.Printf("Error reading reflections directory: %v", err)
			return
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			path := filepath.Join("content/reflections", entry.Name())
			post, err := models.LoadPostFromFS(c.contentFS, path)
			if err != nil {
				log.Printf("Error loading post from %s: %v", path, err)
				continue
			}
			posts = append(posts, *post)
		}

		slices.SortFunc(posts, func(a, b models.Post) int {
			return b.PublishedAt.Compare(a.PublishedAt)
		})

		for i := range posts {
			slugMap[posts[i].Slug] = &posts[i]
		}

		c.allPosts = posts
		c.postBySlug = slugMap
		log.Printf("Loaded %d posts into cache", len(posts))

		// Pre-render posts if render function is set
		if c.renderFuncSet && c.renderFunc != nil {
			c.preRenderAll()
		}

		// Pre-generate RSS feed if RSS generator is set
		if c.rssGenerator != nil {
			c.rssGenerator.Generate(c.allPosts)
		}
	})
}

func (c *PostsCache) preRenderAll() {
	for i := range c.allPosts {
		post := &c.allPosts[i]
		html, err := c.renderFunc(post)
		if err != nil {
			log.Printf("Warning: failed to pre-render post %s: %v", post.Slug, err)
			continue
		}
		if err := post.SetRenderedHTML(html); err != nil {
			log.Printf("Warning: failed to compress post %s: %v", post.Slug, err)
		}
	}
}

// SetContentFS sets the embedded filesystem for reading content files
func (c *PostsCache) SetContentFS(fsys fs.FS) {
	c.contentFS = fsys
}

// SetRenderFunc sets the function used to render posts to HTML
func (c *PostsCache) SetRenderFunc(fn func(*models.Post) ([]byte, error)) {
	c.renderFunc = fn
	c.renderFuncSet = true
}

// SetRSSGenerator sets the RSS generator for feed generation
func (c *PostsCache) SetRSSGenerator(gen *rss.Generator) {
	c.rssGenerator = gen
}

func (c *PostsCache) GetAllPosts() []models.Post {
	c.Init()
	return c.allPosts
}

func (c *PostsCache) GetPostBySlug(slug string) *models.Post {
	c.Init()
	return c.postBySlug[slug]
}
