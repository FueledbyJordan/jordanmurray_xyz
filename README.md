# jordanmurray.xyz

A silly little site.

## Adding New Blog Posts

1. Add a new markdown file `content/reflections` directory
2. Ensure the markdown file has proper metadata
3. Push to master

## Development

Workflow is nix flake driven.  Use `nix develop` to get a development shell.
Use `nix build` to build the application.  To build the container, use `nix build .#container`.
To run the application, use `nix run` or `nix run .#container`.

`tools/gen-chroma-css.go` is used to generate new color schemes for code snippets.
make sure to send the standard output to `static/css/chroma.css`.

## Publish

A container is built and deployed using the `flake.nix` deployment.

## License

MIT
