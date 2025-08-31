package enclave

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"github.com/zeozeozeo/goldmark-enclave/core"
)

type (
	Option           func(*enclaveExtension)
	enclaveExtension struct {
		cfg *core.Config
	}
)

func NewEnclave(c *core.Enclave) *core.Enclave {
	c.Destination = c.Image.Destination
	c.Title = string(c.Image.Title)
	return c
}

func (e *enclaveExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&astTransformer{
				cfg: e.cfg,
			}, 500),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewHTMLRenderer(e.cfg), 500),
		),
	)
}

func New(cfg *core.Config) goldmark.Extender {
	e := &enclaveExtension{
		cfg: cfg,
	}
	return e
}
