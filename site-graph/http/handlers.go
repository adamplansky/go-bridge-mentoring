package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/crawler"

	"go.uber.org/zap"
)

type scrapeParams struct {
	Depth int
	URL   *url.URL
}

// parseScrapeParams parse params from client. If depth in query is empty
// it uses 1 as default parameter. Query parameter url is mandatory and returns
// error if is not specified.
func parseScrapeParams(q url.Values) (*scrapeParams, error) {
	var params scrapeParams
	if qDepth := q.Get("depth"); qDepth != "" {
		depth, err := strconv.Atoi(qDepth)
		if err != nil {
			return nil, fmt.Errorf("query parameter 'depth' is invalid: %w", err)
		}
		params.Depth = depth
	} else {
		params.Depth = 1
	}

	if rawURL := q.Get("url"); rawURL != "" {
		URL, err := url.Parse(rawURL)
		if err != nil {
			return nil, fmt.Errorf("query parameter 'url' is invalid: %w", err)
		}
		params.URL = URL
	} else {
		return nil, fmt.Errorf("query parameter 'url' is empty")
	}

	return &params, nil
}

func (s *server) ScrapeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params, err := parseScrapeParams(r.URL.Query())
	if err != nil {
		s.httpErr(w, http.StatusBadRequest, err)
		return
	}

	collector := crawler.NewCollector(s.log, s.cache)
	g := collector.Work(ctx, *params.URL, params.Depth)
	if len(g.Nodes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.log.Debug("graph output", zap.Int("nodes_number", len(g.Nodes)))

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
