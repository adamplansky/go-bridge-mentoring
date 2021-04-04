package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/multierr"
	"golang.org/x/net/html"
)

type Cache interface {
	// Add insert new key-value into cache, if key already exists
	// in cache returns evicted=true
	Add(key string, value interface{})
	// Get gets value from the coresponding key from cache
	// ok == true if object exist in cache, otherwire ok == false
	Get(key string) (value interface{}, ok bool)
}

type Crawler struct {
	log   *zap.SugaredLogger
	cache Cache
	Graph *Graph
}

func New(log *zap.SugaredLogger, c Cache) *Crawler {
	graph := &Graph{
		Nodes: make([]Node, 0),
		Edges: make(map[Node][]Node, 0),
		mu:    sync.RWMutex{},
	}

	return &Crawler{
		log:   log,
		cache: c,
		Graph: graph,
	}
}

func (c *Crawler) Scrape(ctx context.Context, URL url.URL, maxDepth int) (*Graph, error) {
	err := c.ScrapeRec(ctx, URL, 0, maxDepth)
	if err != nil {
		return nil, err
	}

	return c.Graph, nil
}

func (c *Crawler) ScrapeRec(ctx context.Context, sourceURL url.URL, depth int, maxDepth int) error {
	if depth == maxDepth {
		return nil
	}

	links, err := c.ParseWebsite(ctx, sourceURL)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, target := range links {
		wg.Add(1)
		go func(target link) {
			defer wg.Done()

			if target.IsZero() {
				return
			}

			c.Graph.AddEdge(sourceURL, target.URL())

			if err := c.ScrapeRec(ctx, target.URL(), depth+1, maxDepth); err != nil {
				c.log.Error("website scraping failed",
					zap.Error(err),
					zap.String("website", target.href.String()),
				)
			}
		}(target)
	}
	wg.Wait()

	return nil
}

func (c *Crawler) ScrapeChan(ctx context.Context, jobs <-chan Job, results chan<- Job, errCh chan<- error, maxDepth int) {
	for job := range jobs {
		if job.Depth == maxDepth {
			errCh <- nil
			continue
		}

		job.URL.Path = ""
		job.URL.RawQuery = ""

		links, err := c.ParseWebsite(ctx, job.URL)
		if err != nil {
			errCh <- err
			continue
		}

		for _, target := range links {
			if target.IsZero() {
				continue
			}
			c.Graph.AddEdge(job.URL, target.URL())
			j := Job{
				URL:   target.URL(),
				Depth: job.Depth + 1,
			}
			select {
			case results <- j:
			case <-time.After(2 * time.Second):
				c.log.Errorf("timeout: unable to insert %v to results channel", j.URL)
			}
		}
		errCh <- nil
	}
}

// ParseWebsite parses html url and returns all <a href> elements and returns
// its href values.
func (c *Crawler) ParseWebsite(ctx context.Context, websiteURL url.URL) ([]link, error) {

	webURL := websiteURL.String()
	//g.log.Debug("downloading website",
	//	zap.String("website", websiteURL.Host),
	//)

	if vals, ok := c.cache.Get(webURL); ok {
		links, ok := vals.([]link)
		if !ok {
			return nil, fmt.Errorf("unable to cast values from cache to []links: %s", webURL)
		}
		return links, nil
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, websiteURL.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response is not http.StatusOK, got %s", res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to load html document: %w", err)
	}

	sel := doc.Find("a")
	var errs []error
	var links []link
	for _, n := range sel.Nodes {
		l, err := getAHref(websiteURL, n)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		links = append(links, l)
	}
	c.cache.Add(webURL, links)
	return links, multierr.Combine(errs...)
}

type link struct {
	name string
	href url.URL
}

func (l link) URL() url.URL {
	l.href.Path = ""
	l.href.RawQuery = ""
	return l.href
}

func (l link) IsZero() bool {
	return l.name == "" || l.href.Scheme == "" || l.href.Host == ""
}

// get link struct from <a href="url"></a> if url does not containt
// full URL it is merged with original url.Url. This case applies for relative
// URL without specified host.
func getAHref(original url.URL, a *html.Node) (link, error) {
	if a == nil {
		return link{}, fmt.Errorf("a is nil")
	}

	if a.FirstChild == nil {
		return link{}, fmt.Errorf("a.FirstChild is nil")
	}

	l := link{name: a.FirstChild.Data}
	for _, attr := range a.Attr {
		if attr.Key == "href" {
			u, err := url.Parse(attr.Val)
			if err != nil {
				return l, err
			}
			l.href = *u
		}
	}

	if l.href.Host == "" {
		path := l.href.Path
		l.href = original
		l.href.Path = path
	}

	return l, nil
}
