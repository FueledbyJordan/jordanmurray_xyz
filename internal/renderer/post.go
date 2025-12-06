package renderer

import (
	"bytes"
	"context"

	"jordanmurray.xyz/site/internal/models"
	"jordanmurray.xyz/site/templates"
)

type PostRenderer struct {
	Post models.Post
}

var _ Renderer = PostRenderer{}

func (p PostRenderer) Render(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	component := templates.Reflection(p.Post)
	if err := component.Render(ctx, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
