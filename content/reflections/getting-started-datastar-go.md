---
title: "Getting Started with Datastar and Go"
author: "Jordan Murray"
published_at: 2025-11-30T00:00:00Z
excerpt: "Learn how to build modern, reactive web applications using Datastar and Go."
tags:
  - go
  - datastar
  - web development
---

Datastar is a lightweight framework for building reactive web applications. When combined with Go's powerful backend capabilities, you can create fast, modern web applications with minimal JavaScript.

## Why Datastar?

- **Lightweight**: No heavy JavaScript frameworks needed
- **Reactive**: Build interactive UIs with simple data binding
- **Server-driven**: Keep your logic on the backend where it belongs
- **Progressive Enhancement**: Works without JavaScript, enhanced with it

## Basic Example

Here's a simple counter example using Datastar:

```html
<div data-store="{count: 0}">
  <p>Count: <span data-text="$count"></span></p>
  <button data-on-click="$count++">Increment</button>
</div>
```

This declarative approach keeps your frontend code simple and maintainable. The `data-store` attribute initializes reactive state, `data-text` binds the display, and `data-on-click` handles user interactions.

## Integration with Go

Go's `net/http` package works perfectly with Datastar. You can use templ for type-safe templates that generate the HTML with Datastar attributes.

The beauty of this approach is that you maintain full control over your application logic on the server side, while still providing a reactive user experience.

## Next Steps

- Explore the [Datastar documentation](https://datastar.dev)
- Build your first interactive component
- Learn about server-sent events for real-time updates

Happy coding!
