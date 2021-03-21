package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/crawler"
)

type ScrapeParams struct {
	Depth int
	URL   *url.URL
}

func (p ScrapeParams) Validate() error {
	if p.Depth == 0 {
		return fmt.Errorf("invalid depth in query param")
	}
	if p.URL == nil {
		return fmt.Errorf("invalid url in query param")
	}

	return nil
}

func parseScrapeParams(q url.Values) (*ScrapeParams, error) {
	var params ScrapeParams
	if q.Get("depth") != "" {
		depth, err := strconv.ParseInt(q.Get("depth"), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid depth in query: %w", err)
		}
		params.Depth = int(depth)
	}

	if q.Get("url") != "" {
		URL, err := url.Parse(q.Get("url"))
		if err != nil {
			return nil, fmt.Errorf("invalid url in query: %w", err)
		}
		params.URL = URL
	}
	if err := params.Validate(); err != nil {
		return nil, err
	}

	return &params, nil
}

func ScrapeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		params, err := parseScrapeParams(r.URL.Query())
		if err != nil {
			_ = HttpError(w, 400, err)
			return
		}
		g, err := crawler.Scrape(ctx, *params.URL, params.Depth)
		if err != nil {
			_ = HttpError(w, 500, err)
			return
		}

		b, err := json.Marshal(g)
		if err != nil {
			_ = HttpError(w, 500, err)
			return
		}

		w.Write(b)

	}
}
