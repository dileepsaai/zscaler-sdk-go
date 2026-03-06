package zscaler

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var transportOnce sync.Once

// ZSCALER_SDK_BASE_URL enables endpoint override without API changes.
// Example: http://localhost:8080
func init() {
	transportOnce.Do(func() {
		base := strings.TrimSpace(os.Getenv("ZSCALER_SDK_BASE_URL"))
		if base == "" {
			return
		}
		target, err := url.Parse(base)
		if err != nil || target.Scheme == "" || target.Host == "" {
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

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.target.Scheme
	clone.URL.Host = t.target.Host
	clone.Host = t.target.Host
	return t.base.RoundTrip(clone)
}
