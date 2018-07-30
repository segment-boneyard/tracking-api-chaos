package test

import (
	"testing"

	"net/http"
)

func TestCORS(t *testing.T) {
	cases := []TTData{
		{
			name: "clientCorsIdentify",
			reqFunc: func() *http.Request {
				req := post("/v1/i", `{"userId": "user-id"}`)
				req.Header.Set("Origin", "https://segment.com")
				return req
			},
			headers: http.Header{
				"Access-Control-Allow-Origin": {"https://segment.com"},
				"Content-Type":                {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/i","headers":{"Origin":["https://segment.com"]}}`,
		},
		{
			name: "clientCorsTrack",
			reqFunc: func() *http.Request {
				req := post("/v1/t", `{"event":"Signup"}`)
				req.Header.Set("Origin", "https://example.com")
				return req
			},
			headers: http.Header{
				"Access-Control-Allow-Origin": {"https://example.com"},
				"Content-Type":                {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/t","headers":{"Origin":["https://example.com"]}}`,
		},
	}

	for _, tc := range cases {
		srv := NewServerTest()
		srv.runTestCase(t, tc)
	}
}
