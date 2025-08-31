package enclave

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/zeozeozeo/goldmark-enclave/core"
	"github.com/zeozeozeo/goldmark-enclave/helper"
)

type astTransformer struct {
	cfg *core.Config
}

func (a *astTransformer) InsertFailedHint(n ast.Node, msg string) {
	msgNode := ast.NewString(fmt.Appendf(nil, "\n<!-- goldmark-enclave: %s -->\n", msg))
	msgNode.SetCode(true)
	n.Parent().InsertAfter(n.Parent(), n, msgNode)
}

var audioExtensions = map[string]struct{}{
	".mp3":  {},
	".wav":  {},
	".ogg":  {},
	".m4a":  {},
	".flac": {},
	".aac":  {},
	".weba": {},
	".opus": {},
	".mid":  {},
	".midi": {},
	".aiff": {},
	".aif":  {},
	".wma":  {},
	".ra":   {},
}

func (a *astTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	replaceImages := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if n.Kind() != ast.KindImage {
			return ast.WalkContinue, nil
		}

		img := n.(*ast.Image)
		u, err := url.Parse(string(img.Destination))
		if err != nil {
			a.InsertFailedHint(n, fmt.Sprintf("failed to parse url: %s, %s", img.Destination, err))
			return ast.WalkContinue, nil
		}

		// [alt](url "title")
		// read the title and alt from markdown
		title := string(img.Title)
		altText := helper.ExtractTextRecursivelyByReader(n, reader)

		oid := ""
		theme := "light"
		provider := ""
		params := map[string]string{}
		if u.Host == "www.youtube.com" && u.Path == "/watch" {
			// this is a youtube video: https://www.youtube.com/watch?v={vid}
			provider = core.EnclaveProviderYouTube
			oid = u.Query().Get("v")
		} else if u.Host == "youtu.be" {
			// this is also a youtube video: https://youtu.be/{vid}
			provider = core.EnclaveProviderYouTube
			oid = u.Path[1:]
			oid = strings.Trim(oid, "/")

		} else if u.Host == "www.bilibili.com" && strings.HasPrefix(u.Path, "/video/") {
			// this is a bilibili video: https://www.bilibili.com/video/{vid}
			provider = core.EnclaveProviderBilibili
			oid = u.Path[7:]
			oid = strings.Trim(oid, "/")

		} else if u.Host == "twitter.com" || u.Host == "m.twitter.com" || u.Host == "x.com" {
			// https://twitter.com/{username}/status/{id number}?theme=dark
			provider = core.EnclaveProviderTwitter
			oid = string(img.Destination)
			if u.Host == "x.com" {
				// replace x.com with twitter.com, because x.com doesn't support using x.com as the source host, what a shame
				oid = strings.Replace(oid, "x.com", "twitter.com", 1)
			}
			theme = u.Query().Get("theme")

		} else if u.Host == "tradingview.com" || u.Host == "www.tradingview.com" {
			// https://www.tradingview.com/chart/UC0wWW9o/?symbol=BITFINEX%3ABTCUSD
			provider = core.EnclaveProviderTradingView
			oid = u.Query().Get("symbol")
			theme = u.Query().Get("theme")

		} else if u.Host == "udify.app" || u.Scheme == "dify" {
			// https://udify.app/chatbot/1NaVTsaJ1t54UrNE
			// or
			// dify://udify.app/chatbot/1NaVTsaJ1t54UrNE
			provider = core.EnclaveProviderDifyWidget
			if u.Scheme == "dify" {
				oid = fmt.Sprintf("https://%s", u.Host+u.Path)
			} else {
				oid = string(img.Destination)
			}

		} else if u.Host == "quail.ink" || u.Host == "dev.quail.ink" || u.Host == "quaily.com" {
			// https://quaily.com/{list_slug} or https://quaily.com/{list_slug}/p/{post_slug}
			const re1 = `^([a-zA-Z0-9_-]+)$`
			const re2 = `^([a-zA-Z0-9_-]+)/p/([a-zA-Z0-9_-]+)$`
			if len(u.Path) > 1 {
				p := strings.Trim(u.Path[1:], "/")
				ok1, _ := regexp.MatchString(re1, p)
				ok2, _ := regexp.MatchString(re2, p)
				if ok1 || ok2 {
					provider = core.EnclaveProviderQuailWidget
					oid = string(img.Destination)
					theme = u.Query().Get("theme")
					params["layout"] = u.Query().Get("layout")
				}
			}

		} else if u.Scheme == "quaily" {
			// list: quaily://list/{list_slug}
			// post: quaily://post/{list_slug}/{post_slug}
			// ad: quaily://ads/{ad_uuid}
			reList := regexp.MustCompile(`^/([a-zA-Z0-9_-]+)$`)
			rePost := regexp.MustCompile(`^/([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)$`)

			if u.Host == "list" || u.Host == "post" {
				provider = core.EnclaveProviderQuailWidget
				// find the list slug and post slug using regex
				var listSlug, postSlug string
				var destURL string

				if reList.MatchString(u.Path) {
					matches := reList.FindStringSubmatch(u.Path)
					if len(matches) > 1 {
						listSlug = matches[1]
						destURL = fmt.Sprintf("https://quaily.com/%s", listSlug)
					}
				} else if rePost.MatchString(u.Path) {
					matches := rePost.FindStringSubmatch(u.Path)
					if len(matches) > 2 {
						listSlug = matches[1]
						postSlug = matches[2]
						destURL = fmt.Sprintf("https://quaily.com/%s/p/%s", listSlug, postSlug)
					}
				}

				oid = destURL
				theme = u.Query().Get("theme")
				params["layout"] = u.Query().Get("layout")
			}
		} else if u.Host == "open.spotify.com" {
			// https://open.spotify.com/track/5vdp5UmvTsnMEMESIF2Ym7?si=d4ee09bfd0e941c5
			const re = `^track/([a-zA-Z0-9_-]+)$`
			provider = core.EnclaveProviderSpotify
			if len(u.Path) > 1 {
				p := strings.Trim(u.Path[1:], "/")
				// get the track id after /track/
				ok, _ := regexp.MatchString(re, p)
				if ok {
					oid = strings.Split(p, "/")[1]
				}
			}

		} else if _, ok := audioExtensions[strings.ToLower(filepath.Ext(u.Path))]; ok {
			// this is an audio file
			provider = core.EnclaveHtml5Audio
			oid = string(img.Destination)
		} else {
			// check the resize params
			// form 1: ![](https://example.com/image.jpg?w=200&h=100&center=1)
			// form 2: ![](https://example.com/image.jpg|200x100) or ![](https://example.com/image.jpg|200)
			// form 3: ![alt|200x100](https://example.com/image.jpg) or ![alt|200](https://example.com/image.jpg)
			// if the width and height are numbers only, we assume it's unit is px. If not, we need to parse the unit from the string to check it.
			// supported units: %, px, rem
			w := u.Query().Get("w")
			if w == "" {
				w = u.Query().Get("width")
			}
			h := u.Query().Get("h")
			if h == "" {
				h = u.Query().Get("height")
			}
			align := u.Query().Get("align")

			destination := string(img.Destination)
			reForm := regexp.MustCompile(`\|(\d+%?|rem?|px?)(?:x(\d+%?|rem?|px?))?`)
			// check the form 2, the tail of img.Destination is like |200x100 or |200
			if strings.Contains(destination, "|") {
				matches := reForm.FindStringSubmatch(destination)
				if len(matches) > 1 {
					w = matches[1]
					if len(matches) > 2 {
						h = matches[2]
					}
				}
				destination = strings.Split(destination, "|")[0]
			}

			if strings.Contains(altText, "|") {
				matches := reForm.FindStringSubmatch(altText)
				if len(matches) > 1 {
					w = matches[1]
					if len(matches) > 2 {
						h = matches[2]
					}
				}
			}

			if len(title) != 0 || w != "" || h != "" || align != "" {
				// this is a normal image, but it has a title, so we add a caption
				provider = core.EnclaveProviderQuailImage
				oid = destination
				if title != "" {
					params["title"] = string(img.Title)
				}
				if altText != "" {
					params["alt"] = altText
				}
				if w != "" {
					params["width"] = w
				}
				if h != "" {
					params["height"] = h
				}
				if align != "" {
					params["align"] = align
				}
			} else {
				provider = core.EnclaveRegularImage
				oid = destination
			}
			u, err = url.Parse(destination)
			if err != nil {
				return ast.WalkContinue, nil
			}
		}

		if oid != "" {
			ev := NewEnclave(
				&core.Enclave{
					Image:          *img,
					Alt:            altText,
					Title:          title,
					URL:            u,
					Provider:       provider,
					ObjectID:       oid,
					Theme:          theme,
					Params:         params,
					IframeDisabled: a.cfg.IframeDisabled,
				})

			parent := n.Parent()
			// fmt.Printf("parent: %v\n", parent.Kind())

			// clear the content of the parent node
			// parent.RemoveChildren(parent)
			// add the new enclave node to
			parent.AppendChild(parent, ev)

			// n.Parent().ReplaceChild(n.Parent(), n, ev)

			// for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
			// 	fmt.Printf("child of parent: %+v\n", child)
			// }
		}

		return ast.WalkContinue, nil
	}

	ast.Walk(node, replaceImages)
}
