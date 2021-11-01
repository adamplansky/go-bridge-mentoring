package blackbox

import (
	"net/http"

	"go.uber.org/zap"
)

// Result represent result from scrape operation
type Result struct {
	responses []http.Response
	jobName   string
	err       error
}

// logResult reports results into zap logger.
type logResult struct {
	l *zap.Logger
}

func NewLogResult() (*logResult, error) {
	devLog, err := zap.NewDevelopment()
	return &logResult{
		l: devLog,
	}, err
}

func (l *logResult) Report(resp <-chan Result) {
	for r := range resp {
		if r.err != nil {
			l.l.Error("unable to process request", zap.Error(r.err))
			continue
		}

		for _, resp := range r.responses {
			code := resp.StatusCode
			URL := resp.Request.URL.String()
			l := l.l.With(
				zap.String("URL", URL),
				zap.Int("status", code),
				zap.String("job_name", r.jobName),
			)
			if code >= 200 && code < 300 {
				l.Debug("successful response")
			} else {
				l.Error("response failed")
			}
		}
	}
}
