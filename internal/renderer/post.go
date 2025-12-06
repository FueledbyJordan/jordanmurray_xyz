package renderer

import (
	"bytes"
	"context"
	"fmt"

	"jordanmurray.xyz/site/internal/models"
	"jordanmurray.xyz/site/internal/utils"
	"jordanmurray.xyz/site/templates"
)

type RenderedPost struct {
	models.Post
	HTML           []byte
	CompressedHTML []byte
}

func NewRenderedPost(post models.Post, ctx context.Context) (RenderedPost, error) {
	var buf bytes.Buffer
	component := templates.Reflection(post)
	if err := component.Render(ctx, &buf); err != nil {
		return RenderedPost{}, fmt.Errorf("error rendering post: %w", err)
	}

	renderedHTML := buf.Bytes()
	compressedHTML, err := utils.Compress(renderedHTML, utils.DefaultCompression)
	if err != nil {
		return RenderedPost{}, fmt.Errorf("error compressing post: %w", err)
	}

	return RenderedPost{
		Post:           post,
		HTML:           renderedHTML,
		CompressedHTML: compressedHTML,
	}, nil
}

func (r RenderedPost) Data() []byte {
	return r.HTML
}

func (r RenderedPost) CompressedData() []byte {
	return r.CompressedHTML
}

func (r RenderedPost) ContentType() string {
	return "text/html; charset=utf-8"
}
