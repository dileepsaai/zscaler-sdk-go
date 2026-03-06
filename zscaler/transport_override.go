package zscaler

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var transportOnce sync.Once

// Default mock base used when no env is provided.
// You can override with ZSCALER_SDK_BASE_URL.
const defaultMockBaseURL = "http://192.168.29.160:8080"

// Behavior:
// 1) If ZSCALER_SDK_USE_REAL=true => do not rewrite.
// 2) Else use ZSCALER_SDK_BASE_URL if set.
// 3) Else use defaultMockBaseURL.
func init() {
	transportOnce.Do(func() {
		if strings.EqualFold(strings.TrimSpace(os.Getenv("ZSCALER_SDK_USE_REAL")), "true") {
			return
		}

		base := strings.TrimSpace(os.Getenv("ZSCALER_SDK_BASE_URL"))
		if base == "" {
			base = defaultMockBaseURL
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
