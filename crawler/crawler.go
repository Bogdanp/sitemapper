package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	lane "gopkg.in/oleiade/lane.v1"
)

// Page groups the outward-facing references of a single web page into
// two buckets: links and assets.  Links represent links to other web
// pages, while assets represent links to resources like images, css,
// js, etc.
type Page struct {
	Links  []string
	Assets []string
}

// Sitemap is a mapping from page URLs to Pages.
type Sitemap map[string]Page

// PrettyPrint prints a Sitemap to stdout.
func (s Sitemap) PrettyPrint() {
	for loc, page := range s {
		fmt.Println(loc)
		fmt.Println(" * links: ")
		for _, link := range page.Links {
			fmt.Printf("    - %s\n", link)
		}
		fmt.Println(" * assets: ")
		for _, asset := range page.Assets {
			fmt.Printf("    - %s\n", asset)
		}
		fmt.Println()
	}
}

// Crawler represents a particular web crawler configuration.
type Crawler struct {
	root        *url.URL
	concurrency int

	logger Logger
	client *http.Client
}

// Option is the type of configuration options for the Crawl method.
type Option func(c *Crawler)

// NewCrawler returns a basic Crawler.
func NewCrawler(root *url.URL) *Crawler {
	return &Crawler{
		root:        root,
		concurrency: 8,

		logger: &BasicLogger{},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithConcurrency sets the max concurrency level.
func WithConcurrency(n int) Option {
	return func(c *Crawler) {
		c.concurrency = n
	}
}

// WithLogger sets the logger.
func WithLogger(logger Logger) Option {
	return func(c *Crawler) {
		c.logger = logger
	}
}

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Crawler) {
		c.client = client
	}
}

// Crawl the web page and build a Sitemap.
func (c *Crawler) Crawl(options ...Option) Sitemap {
	for _, option := range options {
		option(c)
	}

	sem := make(chan struct{}, c.concurrency)
	jobs := 0
	operations := lane.NewDeque()
	operations.Append(opPush{nil, c.root})

	links := make(map[string]urls)
	assets := make(map[string]urls)
	visited := make(map[string]bool)

loop:
	for {
		item := operations.Shift()
		if item == nil {
			continue
		}

		switch op := item.(type) {
		case opPop:
			jobs--
			if jobs == 0 {
				break loop
			}

		case opPush:
			c.logger.Debug("Processing page %q referenced by %q...", op.targetURL, op.sourceURL)

			target := op.targetURL.String()
			if visited[target] || op.targetURL.Host != c.root.Host {
				c.logger.Debug("Skipping page %q due to host mismatch...", target)
				continue
			}

			jobs++
			visited[target] = true

			sem <- struct{}{}
			go func() {
				defer func() {
					operations.Append(opPop{})
					<-sem
				}()

				ls, err := processURL(c.client, op.targetURL)
				if err != nil {
					c.logger.Warn("could not process URL %q: %v", op.targetURL, err)
					return
				}

				operations.Append(opTrack{links, op.sourceURL, op.targetURL})
				for _, l := range ls {
					switch l.kind {
					case kindAsset:
						operations.Append(opTrack{assets, op.targetURL, l.url})
					case kindPage:
						operations.Append(opTrack{links, op.targetURL, l.url})
						operations.Append(opPush{op.targetURL, l.url})
					}
				}
			}()

		case opTrack:
			if op.sourceURL == nil {
				continue
			}

			source := op.sourceURL.String()
			target := op.targetURL.String()
			if op.bucket[source] == nil {
				op.bucket[source] = make(urls)
			}

			op.bucket[source][target] = true
		}
	}

	sitemap := make(Sitemap)
	for loc := range visited {
		page := sitemap[loc]
		page.Assets = make([]string, 0, len(assets))
		page.Links = make([]string, 0, len(links))

		for asset := range assets[loc] {
			page.Assets = append(page.Assets, asset)
		}

		for link := range links[loc] {
			page.Links = append(page.Links, link)
		}

		sort.Strings(page.Assets)
		sort.Strings(page.Links)

		sitemap[loc] = page
	}

	return sitemap
}

type urls map[string]bool

type opPop struct{}
type opPush struct {
	sourceURL *url.URL
	targetURL *url.URL
}
type opTrack struct {
	bucket    map[string]urls
	sourceURL *url.URL
	targetURL *url.URL
}

func processURL(client *http.Client, u *url.URL) (links []link, err error) {
	res, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if !strings.HasPrefix(res.Header.Get("content-type"), "text/html") {
		return nil, errors.New("cannot process non-HTML content")
	}

	links, err = findLinks(res.Body, u)
	if err != nil {
		return nil, err
	}

	return links, err
}
