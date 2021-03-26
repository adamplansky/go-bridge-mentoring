package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type scrapeParams struct {
	Depth int
	URL   *url.URL
}

func parseScrapeParams(q url.Values) (*scrapeParams, error) {
	var params scrapeParams
	if q.Get("depth") != "" {
		depth, err := strconv.Atoi(q.Get("depth"))
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
	return &params, nil
}

func (s *server) ScrapeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params, err := parseScrapeParams(r.URL.Query())
	if err != nil {
		httpErr(w, 400, err)
		return
	}
	g, err := s.crawler.Scrape(ctx, *params.URL, params.Depth)
	if err != nil {
		httpErr(w, 500, err)
		return
	}

	err = json.NewEncoder(w).Encode(g)
	if err != nil {
		httpErr(w, 500, err)
		return
	}
}
