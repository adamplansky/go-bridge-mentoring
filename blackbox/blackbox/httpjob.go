package blackbox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type Job struct {
	Name       string        `yaml:"job_name"`
	Targets    []url.URL     `yaml:"targets"`
	Timeout    time.Duration `yaml:"timeout"`
	PeriodTime time.Duration `yaml:"period_time"`
}

func (j Job) IsValid() error {
	if j.Timeout == 0 {
		return fmt.Errorf("job is not valid: timeout is 0")
	}

	if j.PeriodTime == 0 {
		return fmt.Errorf("job is not valid: period time is 0")
	}
	return nil
}

// stopTicker cleans up the channel for garbage collector.
func stopTicker(t *time.Ticker) {
	t.Stop()
	select {
	case <-t.C:
	default:
	}
}

type responseGatherer struct {
	resps []http.Response
	m     sync.Mutex
}

func (r *responseGatherer) add(resp http.Response) {
	r.m.Lock()
	defer r.m.Unlock()
	r.resps = append(r.resps, resp)
}

func (r *responseGatherer) reset() {
	r.resps = r.resps[:0]
}

// Do scrapes all target from job's targets and creates Result from them. This
// Result is passed to scraper reporter using r channel.
func (j Job) Do(ctx context.Context, r chan<- Result) {
	ticker := time.NewTicker(j.PeriodTime)
	defer stopTicker(ticker)
	responses := &responseGatherer{}
	j.scrapeTargets(ctx, r, responses)
	for range ticker.C {
		j.scrapeTargets(ctx, r, responses)
	}
}

func (j Job) scrapeTargets(ctx context.Context, r chan<- Result, responses *responseGatherer) {
	responses.reset()
	ctx, cancel := context.WithTimeout(ctx, j.Timeout)
	g, ctx := errgroup.WithContext(ctx)

	for _, t := range j.Targets {
		scrapedURL := t
		g.Go(func() error {
			resp, err := httpScrape(ctx, scrapedURL, j.Timeout)
			if err != nil {
				return err
			}
			responses.add(*resp)
			return nil
		})
	}
	err := g.Wait()
	cancel()

	select {
	case <-ctx.Done():
		break
	case r <- Result{responses: responses.resps, err: err, jobName: j.Name}:

	}
}

// httpScrape process HTTP.GET request for url.
func httpScrape(ctx context.Context, url url.URL, timeout time.Duration) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}
	c := http.Client{
		Timeout: timeout,
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client Do failed: %w", err)
	}
	defer resp.Body.Close()

	return resp, nil
}
