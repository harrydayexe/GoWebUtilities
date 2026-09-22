package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectWWW(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := NewRedirectWWW()(next)

	cases := []struct {
		name         string
		url          string
		wantStatus   int
		wantLocation string
	}{
		{
			name:         "www root redirects to apex",
			url:          "https://www.example.com/",
			wantStatus:   http.StatusMovedPermanently,
			wantLocation: "https://example.com/",
		},
		{
			name:         "www keeps path and query",
			url:          "https://www.example.com/blog/post?ref=x&y=1",
			wantStatus:   http.StatusMovedPermanently,
			wantLocation: "https://example.com/blog/post?ref=x&y=1",
		},
		{
			name:         "www over http redirects to https apex",
			url:          "http://www.example.com/about",
			wantStatus:   http.StatusMovedPermanently,
			wantLocation: "https://example.com/about",
		},
		{
			name:       "apex passes through",
			url:        "https://example.com/",
			wantStatus: http.StatusOK,
		},
		{
			name:       "other subdomain passes through",
			url:        "https://api.example.com/",
			wantStatus: http.StatusOK,
		},
		{
			name:       "host starting with www but no dot passes through",
			url:        "https://wwwexample.com/",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if got := rec.Header().Get("Location"); got != tc.wantLocation {
				t.Errorf("Location = %q, want %q", got, tc.wantLocation)
			}
		})
	}
}
