package webutil

import (
	"net/http"
	"time"
)

type authTransport struct {
	http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+t.token)
	return t.RoundTripper.RoundTrip(r)
}

func NewClientWithAuth(token string, timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &authTransport{
			RoundTripper: http.DefaultTransport,
			token:        token,
		},
	}
}
