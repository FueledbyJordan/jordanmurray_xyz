# Jordan Murray's Blog

A modern blog built with Go, Templ, Datastar, and DaisyUI.

## Tech Stack

- **Go**: Backend server and routing
- **Templ**: Type-safe HTML templating
- **Datastar**: Lightweight reactivity and interactivity
- **DaisyUI**: Beautiful Tailwind CSS components

## Prerequisites

- Go 1.21 or higher
- Templ CLI (for template generation)

## Installation

1. Clone the repository:
```bash
cd jordanmurray_xyz
```

2. Install templ CLI:
```bash
go install github.com/a-h/templ/cmd/templ@latest
```

3. Install Go dependencies:
```bash
go mod tidy
```

4. Generate templ templates:
```bash
templ generate
```

## Running the Application

1. Generate templates (if not done already):
```bash
templ generate
```

2. Run the server:
```bash
go run main.go
```

3. Open your browser to [http://localhost:8080](http://localhost:8080)

## Development Workflow

When you make changes to `.templ` files, you need to regenerate them:

```bash
# Watch mode (regenerates automatically on changes)
templ generate --watch

# Or run once
templ generate
```

Then restart the Go server:
```bash
go run main.go
```

## Project Structure

```
.
├── main.go              # Application entry point
├── handlers/            # HTTP handlers
│   └── handlers.go
├── models/              # Data models
│   └── post.go
├── templates/           # Templ templates
│   ├── layout.templ     # Base layout with DaisyUI
│   ├── home.templ       # Homepage
│   ├── blog.templ       # Blog list and post views
│   └── datastar_example.templ  # Datastar examples
├── content/             # Blog post content
│   └── posts/
├── static/              # Static assets
│   ├── css/
│   └── js/
└── go.mod               # Go module definition
```

## Features

- Responsive design with DaisyUI components
- Server-side rendering with Templ
- Interactive UI elements with Datastar
- Blog post listing and detail views
- Clean, modern design

## Adding New Blog Posts

Blog posts are currently defined in `models/post.go`. To add a new post:

1. Add a new `Post` struct to the `GetAllPosts()` function
2. Create corresponding markdown content in `content/posts/` (for reference)

Future enhancement: Load posts from markdown files dynamically.

## Customization

### Changing the Theme

Edit `templates/layout.templ` and change the `data-theme` attribute:

```html
<html lang="en" data-theme="dark">
```

Available themes: light, dark, cupcake, bumblebee, emerald, corporate, synthwave, retro, cyberpunk, valentine, halloween, garden, forest, aqua, lofi, pastel, fantasy, wireframe, black, luxury, dracula, cmyk, autumn, business, acid, lemonade, night, coffee, winter, dim, nord, sunset

### Adding Datastar Interactivity

Use Datastar attributes in your templ templates:

```templ
<div data-store="{count: 0}">
  <span data-text="$count"></span>
  <button data-on-click="$count++">Increment</button>
</div>
```

## Resources

- [Templ Documentation](https://templ.guide/)
- [Datastar Documentation](https://data-star.dev/)
- [DaisyUI Documentation](https://daisyui.com/)
- [Go Documentation](https://golang.org/doc/)

## License

MIT
