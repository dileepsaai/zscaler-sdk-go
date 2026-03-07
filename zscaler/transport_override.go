package zscaler

import (
	"net/http"
	"net/url"
	"sync"
)

var transportOnce sync.Once

// Hardcoded API server URL (auth + resource calls are rewritten here).
const hardcodedAPIBaseURL = "https://highborn-nonconsecutive-velva.ngrok-free.dev"

// GetMockTarget returns the hardcoded API target.
func GetMockTarget() (*url.URL, bool) {
	target, err := url.Parse(hardcodedAPIBaseURL)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return nil, false
	}
	return target, true
}

func init() {
	transportOnce.Do(func() {
		target, ok := GetMockTarget()
		if !ok || target == nil {
			return
		}
		baseRT, ok := http.DefaultTransport.(*http.Transport)
		if !ok || baseRT == nil {
			return
		}
		clone := baseRT.Clone()
		http.DefaultTransport = &rewriteTransport{target: target, base: clone}
	})
}

type rewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func NewRewriteTransport(base http.RoundTripper) http.RoundTripper {
	target, ok := GetMockTarget()
	if !ok || target == nil {
		return base
	}
	return &rewriteTransport{target: target, base: base}
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.target.Scheme
	clone.URL.Host = t.target.Host
	clone.Host = t.target.Host
	return t.base.RoundTrip(clone)
}
