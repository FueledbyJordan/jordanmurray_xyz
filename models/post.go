package models

import "time"

type Post struct {
	ID          string
	Title       string
	Slug        string
	Author      string
	PublishedAt time.Time
	Content     string
	Excerpt     string
	Tags        []string
}

// Mock data for demonstration
func GetAllPosts() []Post {
	return []Post{
		{
			ID:          "1",
			Title:       "Getting Started with Datastar and Go",
			Slug:        "getting-started-datastar-go",
			Author:      "Jordan Murray",
			PublishedAt: time.Now().AddDate(0, 0, -2),
			Excerpt:     "Learn how to build modern, reactive web applications using Datastar and Go.",
			Content:     "Datastar is a lightweight framework for building reactive web applications...",
			Tags:        []string{"go", "datastar", "web development"},
		},
		{
			ID:          "2",
			Title:       "Styling with DaisyUI",
			Slug:        "styling-with-daisyui",
			Author:      "Jordan Murray",
			PublishedAt: time.Now().AddDate(0, 0, -5),
			Excerpt:     "Discover how DaisyUI makes Tailwind CSS even more productive.",
			Content:     "DaisyUI provides beautiful, ready-to-use components built on Tailwind CSS...",
			Tags:        []string{"css", "tailwind", "daisyui"},
		},
	}
}

func GetPostBySlug(slug string) *Post {
	posts := GetAllPosts()
	for _, post := range posts {
		if post.Slug == slug {
			return &post
		}
	}
	return nil
}
