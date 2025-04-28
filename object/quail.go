package object

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/quailyquaily/goldmark-enclave/core"
)

const quailWidgetTpl = `
<iframe
	src="{{.URL}}"
	data-theme="{{.Theme}}"
	width="100%"
	height="{{.Height}}"
	title="Quail Widget"
	frameborder="0"
	allow="web-share"
	allowfullscreen
></iframe>
`

const quailImageTpl = `
<figure class="quail-image-wrapper" style="width: {{.Width}}; height: {{.Height}}; margin: 1rem 0; display: block">
	<img src="{{.URL}}" alt="{{.Alt}}" style="width: {{.Width}}; height: {{.Height}}" class="quail-image" />
	<figcaption class="quail-image-caption" style="display: block">{{.Title}}</figcaption>
</figure>
`

const quailAdTpl = `
<div class="quail-ad-wrapper" style="width: 100%; height: auto; margin: 1rem 0; display: block">
	<div class="quail-ad" data-ad-uuid="{{.ObjectID}}" style="width: 100%; height: auto"></div>
</div>
`

func GetQuailWidgetHtml(enc *core.Enclave) (string, error) {
	if enc.Theme == "dark" {
		enc.Theme = "dark"
	} else {
		enc.Theme = "light"
	}
	var err error

	ret := ""
	buf := bytes.Buffer{}
	if enc.IframeDisabled {
		ret, err = GetNoIframeTplHtml(enc, fmt.Sprintf("%s://%s%s", enc.URL.Scheme, enc.URL.Host, enc.URL.Path))
		if err != nil {
			return "", err
		}

	} else {
		t, err := template.New("quail-widget").Parse(quailWidgetTpl)
		if err != nil {
			return "", err
		}

		layout := ""
		if l, ok := enc.Params["layout"]; ok {
			layout = l
		}

		height := "auto"
		if strings.Contains(enc.URL.Path, "/p/") {
			height = "128px"
		} else if layout == "subscribe_form" {
			height = "390px"
		} else if layout == "subscribe_form_mini" {
			height = "142px"
		}

		if err = t.Execute(&buf, map[string]string{
			"URL":    fmt.Sprintf("%s://%s%s/widget?theme=%s&layout=%s&logged=ignore", enc.URL.Scheme, enc.URL.Host, enc.URL.Path, enc.Theme, layout),
			"Theme":  enc.Theme,
			"Height": height,
		}); err != nil {
			return "", err
		}

		ret = buf.String()
	}

	return ret, nil
}

func GetQuailImageHtml(enc *core.Enclave) (string, error) {
	buf := bytes.Buffer{}

	t, err := template.New("quail-image").Parse(quailImageTpl)
	if err != nil {
		return "", err
	}

	w := "auto"
	if width, ok := enc.Params["width"]; ok {
		w = width
	}

	h := "auto"
	if height, ok := enc.Params["height"]; ok {
		h = height
	}

	// if w and h are number, add px to the end
	if wNum, err := strconv.Atoi(w); err == nil {
		w = fmt.Sprintf("%dpx", wNum)
	}
	if hNum, err := strconv.Atoi(h); err == nil {
		h = fmt.Sprintf("%dpx", hNum)
	}

	title := enc.Title
	alt := enc.Alt

	if err = t.Execute(&buf, map[string]string{
		"URL":    enc.URL.String(),
		"Title":  title,
		"Width":  w,
		"Height": h,
		"Alt":    alt,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func GetQuailAdHtml(enc *core.Enclave) (string, error) {
	buf := bytes.Buffer{}

	t, err := template.New("quail-ad").Parse(quailAdTpl)
	if err != nil {
		return "", err
	}

	if err = t.Execute(&buf, map[string]string{
		"ObjectID": enc.ObjectID,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
