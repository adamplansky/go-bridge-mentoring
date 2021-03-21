package roundtripper

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type debugTransport struct {
	l  Logger
	rt http.RoundTripper
}

type Logger interface {
	Debugf(format string, args ...interface{})
}

func NewDebug(rt http.RoundTripper, l Logger) http.RoundTripper {
	return &debugTransport{
		rt: rt,
		l:  l,
	}
}

func (dt *debugTransport) RoundTrip(r *http.Request) (*http.Response, error) {

	b, err := httputil.DumpRequest(r, false)
	if err != nil {
		return nil, fmt.Errorf("unable to dump rt request: %w", err)
	}

	dt.l.Debugf("request: %v", string(b))

	resp, err := dt.rt.RoundTrip(r)
	if err != nil {
		return nil, fmt.Errorf("unable to RoundTrip rt: %w", err)
	}
	b, err = httputil.DumpResponse(resp, false)
	if err != nil {
		return nil, fmt.Errorf("unable to dump rt response: %w", err)
	}
	dt.l.Debugf("response: %v", string(b))

	return resp, nil
}
