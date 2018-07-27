package test

import (
	"net/http/httptest"
	"testing"

	"net/http"

	"github.com/bmizerany/assert"
)

func TestHealth(t *testing.T) {
	cases := []TTData{
		{
			name: "health",
			req:  get("/internal/health", ``),
			headers: http.Header{
				"Content-Length": {"0"},
				"Content-Type":   {"text/plain"},
			},
			code:     http.StatusOK,
			bodyResp: ``,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, td := NewServerTest()
			defer td()

			rec := httptest.NewRecorder()
			req := tc.req
			if req == nil {
				req = tc.reqFunc()
			}
			srv.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tc.code)
			assert.Equal(t, rec.Body.String(), tc.bodyResp)
			for k, v := range tc.headers {
				assert.Equal(t, v[0], rec.Header().Get(k))
			}
		})
	}
}
