package crawler

import (
	"context"
	"net/url"

	"go.uber.org/zap"
)

type Collector struct {
	queue   []string
	crawler *Crawler
	log     *zap.SugaredLogger
}

func NewCollector(log *zap.SugaredLogger, c Cache) *Collector {
	crawler := New(log, c)
	return &Collector{
		crawler: crawler,
		queue:   make([]string, 0),
		log:     log,
	}
}

type Job struct {
	URL   url.URL
	Depth int
}

func (c *Collector) Work(ctx context.Context, URL url.URL, maxDepth int) *Graph {
	const limit = 10
	jobs := make(chan Job, limit)
	results := make(chan Job, limit)
	errCh := make(chan error)

	for i := 0; i < limit; i++ {
		go c.crawler.ScrapeChan(ctx, jobs, results, errCh, maxDepth)
	}

	// init jobs
	jobsCount := 1
	jobs <- Job{
		URL:   URL,
		Depth: 0,
	}
	// end init jobs

	visitedSites := make(map[url.URL]bool, 0)
	queueSites := make([]Job, 0)
	for jobsCount > 0 {
		select {
		// results channel returns newJob (new links to parse)
		case newJob := <-results:
			// handle enqueue - add only not parsed sites
			if ok := visitedSites[newJob.URL]; !ok {
				jobsCount++
				visitedSites[newJob.URL] = true
				queueSites = append(queueSites, newJob)
			}
		// errCh channel is for signaling successfully or unsuccessfully parsed site
		case err := <-errCh:
			if err != nil {
				c.log.Errorf("crawler.ScrapeChan: %s", err)
			}

			jobsCount--
			// we have a guarantee there is a space in jobs channel
			// because of <-errCh
			if len(queueSites) > 0 {
				jobs <- queueSites[0]
				// remove first element from slice
				queueSites = append(queueSites[:0], queueSites[1:]...)
			}

		}
	}
	close(jobs)
	close(results)
	close(errCh)

	return c.crawler.Graph
}
