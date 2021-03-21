package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/multierr"
	"golang.org/x/net/html"
)

//var ErrQueueEmpty = errors.New("queue is empty")

func Scrape(ctx context.Context, websiteURL url.URL, maxDepth int) (*Graph, error) {
	g := Graph{
		queueNode: make(map[url.URL]Status),
		Nodes:     make([]Node, 0),
		Edges:     make([]Edge, 0),
	}

	err := g.ScrapeRec(ctx, websiteURL, 0, maxDepth)
	if err != nil {
		return nil, err
	}

	for _, edge := range g.Edges {
		fmt.Println(edge)
	}

	return &g, nil
}

type Status int

const (
	None Status = iota
	Quoted
	InProgress
	Completed
	Failed
)

type Node struct {
	ID string `json:"id"`
}

type Graph struct {
	queueNode map[url.URL]Status
	Nodes     []Node
	Edges     []Edge
}

func (g *Graph) IsCompleted(websiteURL url.URL) bool {
	return g.queueNode[websiteURL] == Completed
}

func (g *Graph) IsInProgress(websiteURL url.URL) bool {
	return g.queueNode[websiteURL] == InProgress
}

func (g *Graph) IsQuoted(websiteURL url.URL) bool {
	return g.queueNode[websiteURL] == Quoted || g.queueNode[websiteURL] == None
}

func (g *Graph) AddNode(websiteURL url.URL) {
	wURL := websiteURL.String()
	for _, n := range g.Nodes {
		if n.ID == wURL {
			return
		}
	}
	g.Nodes = append(g.Nodes, Node{
		ID: wURL,
	})
}

func (g *Graph) ScrapeRec(ctx context.Context, sourceURL url.URL, depth int, maxDepth int) error {
	if depth == maxDepth {
		return nil
	}

	links, _ := ParseWebsite(ctx, sourceURL)
	for _, target := range links {
		if target.IsZero() {
			continue
		}
		// do not traverse all subpages - only host
		// maybe it will be extended in the future
		target.href.Path = ""
		target.href.RawQuery = ""

		// if URL is already parsed skip it
		if !g.IsQuoted(target.href) {
			continue
		}

		fmt.Println(depth, strings.TrimSpace(target.name), target.href.String())
		g.Edges = append(g.Edges, Edge{
			Source: sourceURL,
			Target: target.href,
			Type:   "link",
		})

		// Add sourceURL / targetURL
		g.AddNode(sourceURL)
		g.AddNode(target.href)

		g.queueNode[target.href] = InProgress
		if err := g.ScrapeRec(ctx, target.href, depth+1, maxDepth); err != nil {
			return err
		}

	}

	fmt.Println("source URL: ", sourceURL)
	return nil
}

type Edge struct {
	Source url.URL
	Target url.URL
	Type   string
}

func (e Edge) String() string {
	return fmt.Sprintf("%v -> %v, type: %v\n", e.Source.String(), e.Target.String(), e.Type)
}

type link struct {
	name string
	href url.URL
}

func (l link) IsZero() bool {
	return l.name == "" || l.href.Scheme == "" || l.href.Host == ""
}

// ParseWebsite parses html url and returns all <a href> elements and returns
// its href values.
func ParseWebsite(ctx context.Context, websiteURL url.URL) ([]link, error) {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, websiteURL.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
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
		//fmt.Printf("%#v\n", l)
	}
	return links, multierr.Combine(errs...)
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
