# Getting Started with Datastar and Go

Datastar is a lightweight framework for building reactive web applications. When combined with Go's powerful backend capabilities, you can create fast, modern web applications with minimal JavaScript.

## Why Datastar?

- **Lightweight**: No heavy JavaScript frameworks needed
- **Reactive**: Build interactive UIs with simple data binding
- **Server-driven**: Keep your logic on the backend where it belongs
- **Progressive Enhancement**: Works without JavaScript, enhanced with it

## Basic Example

Here's a simple counter example:

```html
<div data-store="{count: 0}">
  <p>Count: <span data-text="$count"></span></p>
  <button data-on-click="$count++">Increment</button>
</div>
```

## Integration with Go

Go's net/http package works perfectly with Datastar. You can use templ for type-safe templates that generate the HTML with Datastar attributes.

## Next Steps

- Explore the Datastar documentation
- Build your first interactive component
- Learn about server-sent events for real-time updates

Happy coding!
