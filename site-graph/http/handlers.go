package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go.uber.org/zap"
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
		s.httpErr(w, 400, err)
		return
	}
	g, err := s.crawler.Scrape(ctx, *params.URL, params.Depth)
	if err != nil {
		s.httpErr(w, 500, err)
		return
	}

	if len(g.Nodes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.log.Debug("graph output", zap.Int("nodes_number", len(g.Nodes)))
	fmt.Println()

	s.resp(w, g)

}

func (s *server) resp(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		s.httpErr(w, http.StatusInternalServerError, err)
		return
	}
}
