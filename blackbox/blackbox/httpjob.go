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

func (j Job) Do(r chan<- Result) {
	var mu sync.Mutex
	var responses []http.Response
	ticker := time.Tick(j.PeriodTime)

	for range ticker {
		responses = make([]http.Response, 0, len(j.Targets))
		ctx, cancel := context.WithTimeout(context.Background(), j.Timeout)
		g, ctx := errgroup.WithContext(ctx)
		for _, t := range j.Targets {
			scrapedURL := t
			g.Go(func() error {
				resp, err := httpScrape(ctx, scrapedURL, j.Timeout)
				if err != nil {
					return err
				}
				mu.Lock()
				responses = append(responses, *resp)
				mu.Unlock()
				return nil
			})
		}
		err := g.Wait()
		cancel()

		r <- Result{
			responses: responses,
			err:       err,
			jobName:   j.Name,
		}
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

	return resp, nil
}
