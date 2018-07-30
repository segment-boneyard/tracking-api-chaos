package test

import (
	"testing"

	"net/http"
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
		srv := NewServerTest()
		srv.runTestCase(t, tc)
	}
}
