package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"
)

func ParseURL(rawurl string) url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return *u
}

func main() {
	start := time.Now()
	scrapes := []Scrapes{
		{
			JobName: "testCase1",
			Timeout: time.Second * 15,
			Targets: []url.URL{
				ParseURL("https://prometheus.io/"),
				ParseURL("https://www.google.com/"),
				ParseURL("https://grafana.com/"),
				ParseURL("https://github.com/kubernetes/kubernetes"),
				ParseURL("https://kubernetes.io/"),
				ParseURL("https://tailscale.com/"),
			},
		},
		{
			JobName: "testCase2",
			Timeout: time.Second * 15,
			Targets: []url.URL{
				ParseURL("https://au.finance.yahoo.com/"),
				ParseURL("https://twitter.com/home"),
			},
		},
	}
	done := make(chan bool)
	jobs := make(chan Scrapes, 5)
	result := make(chan Result)

	var r reporter = &logResult{}

	go scrapeJob(jobs, done, result)
	go r.Report(result)

	for {
		// how to implement scraping in periods? every job can have
		// different periodSeconds - is it possible to implement it somehow
		// as select and time.After?
		for _, scrape := range scrapes {
			jobs <- scrape
		}
		time.Sleep(time.Second * 5)
	}
	// how to handle closing jobs on infinite jobs as this leads into unreachable code
	close(jobs)

	<-done
	fmt.Println("done!")
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println(elapsed)
}

// scrape receive job and spin up new go routine for each scrapeURL for job.Target.
// Resulting responses are sent into resulsts channel.
func scrape(job Scrapes, results chan Result) error {
	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	for _, t := range job.Targets {
		scrapedURL := t
		g.Go(func() error {
			// Fetch the URL.
			resp, err := scrapeURL(context.Background(), scrapedURL)
			if err != nil {
				return err
			}
			results <- Result{resp: resp}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	fmt.Println("Successfully fetched all URLs.")
	return nil

}

// scrape job receive job
func scrapeJob(jobs chan Scrapes, done chan bool, results chan Result) {
	defer close(results)
	for {
		job, ok := <-jobs
		if ok {
			fmt.Println("received job", job)
			err := scrape(job, results)
			if err != nil {
				results <- Result{err: err}
			}

		} else {
			fmt.Println("received all jobs, closing channel")
			done <- true
			return
		}
	}
}

// scrapeURL create HTTP.GET request for url.
func scrapeURL(ctx context.Context, url url.URL) (http.Response, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return http.Response{}, fmt.Errorf("new request failed: %w", err)
	}
	c := http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := c.Do(r)
	if err != nil {
		return http.Response{}, fmt.Errorf("client Do failed: %w", err)
	}

	return *resp, nil
}

// ScrapeConfigs
type Scrapes struct {
	JobName string        `yaml:"job_name"`
	Targets []url.URL     `yaml:"targets"`
	Timeout time.Duration `yaml:"timeout"`
}

// Result represent result from scrape operation
type Result struct {
	resp http.Response
	err  error
}

// reporter reports results from every scrape target
type reporter interface {
	Report(resp <-chan Result)
}

type logResult struct{}

func (l *logResult) Report(resps <-chan Result) {
	for r := range resps {
		resp := r.resp
		code := resp.StatusCode
		u := resp.Request.URL.String()
		if code >= 200 && code < 300 {
			fmt.Println("successful response: ", u)
		} else {
			fmt.Println("response failed: ", u)
		}
	}
}
