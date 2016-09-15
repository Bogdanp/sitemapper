package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/Bogdanp/sitemapper/crawler"
)

var start string
var concurrency int
var verbose bool
var cpuprofile string

func init() {
	flag.IntVar(&concurrency, "concurrency", runtime.NumCPU(), "The maximum number of concurrent requests to run.")
	flag.StringVar(&start, "start", "http://example.com", "The URL to start crawling.")
	flag.BoolVar(&verbose, "verbose", false, "Print debug info.")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "Write CPU profile.")
}

func assert(err error, message string, params ...interface{}) {
	if err != nil {
		log.Fatalf("error: "+message, params...)
	}
}

func main() {
	flag.Parse()

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		assert(err, "failed to create cpuprifle file: %v", err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	root, err := url.Parse(start)
	assert(err, "%q is not a valid URL", start)

	var logger crawler.Logger
	if verbose {
		logger = &crawler.BasicLogger{}
	} else {
		logger = &crawler.NoopLogger{}
	}

	c := crawler.NewCrawler(root)
	sitemap := c.Crawl(
		crawler.WithConcurrency(concurrency),
		crawler.WithLogger(logger),
	)
	sitemap.PrettyPrint()
}
