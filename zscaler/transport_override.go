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
const defaultMockBaseURL = "https://506111aaff603a.lhr.life"

// GetMockTarget returns the URL to rewrite requests to when mock mode is active.
// Mock mode is on when ZSCALER_SDK_USE_REAL is not "true". Returns (target, true) in mock mode,
// or (nil, false) when using real Zscaler APIs. Used by OneAPI HTTP clients so they hit your mock server.
func GetMockTarget() (*url.URL, bool) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("ZSCALER_SDK_USE_REAL")), "true") {
		return nil, false
	}
	base := strings.TrimSpace(os.Getenv("ZSCALER_SDK_BASE_URL"))
	if base == "" {
		base = defaultMockBaseURL
	}
	target, err := url.Parse(base)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return nil, false
	}
	return target, true
}

// Behavior:
// 1) If ZSCALER_SDK_USE_REAL=true => do not rewrite.
// 2) Else use ZSCALER_SDK_BASE_URL if set.
// 3) Else use defaultMockBaseURL.
func init() {
	transportOnce.Do(func() {
		target, useMock := GetMockTarget()
		if !useMock || target == nil {
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

// RewriteTransport rewrites the request host/scheme to a target URL (e.g. mock server); path is unchanged.
type rewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

// NewRewriteTransport wraps base so that requests are sent to the mock target when mock mode is on.
// When not in mock mode, returns base unchanged.
func NewRewriteTransport(base http.RoundTripper) http.RoundTripper {
	target, useMock := GetMockTarget()
	if !useMock || target == nil {
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
