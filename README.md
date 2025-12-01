# jordanmurray.xyz

A site built with Go, Templ, Datastar, and DaisyUI.

## Adding New Blog Posts

Blog posts are currently defined in `models/post.go`. To add a new post:

1. Add a new `Post` struct to the `GetAllPosts()` function
2. Create corresponding markdown content in `content/posts/` (for reference)

Future enhancement: Load posts from markdown files dynamically.

## License

MIT
