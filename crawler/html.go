package crawler

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type linkKind int

const (
	kindAsset linkKind = iota
	kindPage
)

type link struct {
	kind linkKind
	url  *url.URL
}

func findAttr(n *html.Node, k string) string {
	for _, a := range n.Attr {
		if a.Key == k {
			return a.Val
		}
	}

	return ""
}

func findLinks(in io.Reader, root *url.URL) ([]link, error) {
	var links []link

	doc, err := html.Parse(in)
	if err != nil {
		return nil, err
	}

	n := doc
	for {
		if n.Type == html.ElementNode {
			var loc string
			var kind linkKind

			switch n.Data {
			case "a":
				loc = findAttr(n, "href")
				if strings.HasPrefix(loc, "#") {
					loc = ""
				}

				kind = kindPage
			case "img", "script":
				loc = findAttr(n, "src")
				kind = kindAsset
			case "link":
				loc = findAttr(n, "href")
				kind = kindAsset
			}

			if loc != "" {
				u, err := url.Parse(strings.TrimSpace(loc))
				if err == nil {
					if !u.IsAbs() {
						u = root.ResolveReference(u)
					}

					links = append(links, link{kind, u})
				}
			}
		}

		if n.FirstChild != nil {
			n = n.FirstChild
		} else {
			for n.NextSibling == nil {
				if n == doc {
					return links, nil
				}

				n = n.Parent
			}

			n = n.NextSibling
		}
	}
}
