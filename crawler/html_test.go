package crawler

import (
	"bytes"
	"net/url"
	"testing"
)

type testLink struct {
	kind linkKind
	url  string
}

var tests = []struct {
	input string
	links []testLink
}{
	{
		input: "",
		links: []testLink{},
	},

	{
		input: `
<html>
  <head>
    <link rel="stylesheet" type="text/css" href="http://example.com/base.css" />
    <link rel="stylesheet" type="text/css" href="../buttons.css" />
  </head>
  <body>
    <h1>Some web page</h1>
    <a href="https://golang.org"><img src="https://golang.org/doc/gopher/frontpage.png"/></a>
    <script src="scripts.js"></script>
  </body>
</html>
`,
		links: []testLink{
			{kindAsset, "http://example.com/base.css"},
			{kindAsset, "http://example.com/buttons.css"},
			{kindPage, "https://golang.org"},
			{kindAsset, "https://golang.org/doc/gopher/frontpage.png"},
			{kindAsset, "http://example.com/scripts.js"},
		},
	},
}

func TestFindLinks(t *testing.T) {
	for ti, test := range tests {
		buf := bytes.NewBufferString(test.input)

		root, _ := url.Parse("http://example.com")
		links, err := findLinks(buf, root)
		if err != nil {
			t.Fatalf("findLinks: %v", err)
		}

		if len(test.links) != len(links) {
			t.Errorf("expected %d links, got %d", len(test.links), len(links))
		}

		for i, link := range links {
			v := test.links[i]
			if v.kind != link.kind || v.url != link.url.String() {
				t.Errorf("expected %+v at position %d in test %d, got {kind:%d url:%s}", v, i, ti, link.kind, link.url.String())
			}
		}
	}
}
