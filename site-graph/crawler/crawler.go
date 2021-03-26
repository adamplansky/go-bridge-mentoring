package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/multierr"
	"golang.org/x/net/html"
)

//var ErrQueueEmpty = errors.New("queue is empty")

type Crawler struct {
	log *zap.SugaredLogger
}

func New(log *zap.SugaredLogger) *Crawler {
	return &Crawler{
		log: log,
	}
}

func (c *Crawler) Scrape(ctx context.Context, websiteURL url.URL, maxDepth int) (*Graph, error) {
	g := Graph{
		queueNode: make(map[url.URL]Status),
		Nodes:     make([]Node, 0),
		Edges:     make([]Edge, 0),
		log:       c.log,
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

var _ json.Marshaler = &Node{}

type Node struct {
	ID    url.URL `json:"id"`
	Group int     `json:"group"`
}

func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID    string `json:"id"`
		Group int    `json:"group"`
	}{
		ID:    n.ID.Host,
		Group: n.Group,
	})
}

type Graph struct {
	queueNode map[url.URL]Status
	Nodes     []Node `json:"nodes"`
	Edges     []Edge `json:"links"`
	log       *zap.SugaredLogger
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
	for _, n := range g.Nodes {
		if n.ID == websiteURL {
			return
		}
	}
	g.Nodes = append(g.Nodes, Node{
		ID:    websiteURL,
		Group: 1,
	})
}

func (g *Graph) ScrapeRec(ctx context.Context, sourceURL url.URL, depth int, maxDepth int) error {
	if depth == maxDepth {
		return nil
	}

	links, err := ParseWebsite(ctx, sourceURL)
	if err != nil {
		return err
	}
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

		//fmt.Println(depth, strings.TrimSpace(target.name), target.href.String())
		g.Edges = append(g.Edges, Edge{
			Source: sourceURL,
			Target: target.href,
			Type:   "link",
			Value:  1,
		})

		// Add sourceURL / targetURL
		g.AddNode(sourceURL)
		g.AddNode(target.href)

		g.queueNode[target.href] = InProgress
		if err := g.ScrapeRec(ctx, target.href, depth+1, maxDepth); err != nil {
			g.log.Error("website scraping failed",
				zap.Error(err),
				zap.String("website", target.href.String()),
			)
			// if error occurs continue scraping
			continue
		}

	}

	//fmt.Println("source URL: ", sourceURL)
	return nil
}

type Edge struct {
	Source url.URL
	Target url.URL
	Type   string
	Value  int
}

func (e *Edge) String() string {
	return fmt.Sprintf("%v -> %v, type: %v\n", e.Source.String(), e.Target.String(), e.Type)
}

func (e *Edge) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Type   string `json:"type"`
		Value  int    `json:"value"`
	}{
		Source: e.Source.Host,
		Target: e.Target.Host,
		Type:   e.Type,
		Value:  e.Value,
	})
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
