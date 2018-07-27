package test

import (
	"net/http/httptest"
	"testing"

	"net/http"

	"github.com/bmizerany/assert"
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
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/i","headers":{"Origin":["https://segment.com"]}}`,
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
			nsqResp:  `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/t","headers":{"Origin":["https://example.com"]}}`,
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

			msg, err := srv.consume()
			assert.Equal(t, err, nil)
			assert.Equal(t, string(msg.Body), tc.nsqResp)
		})
	}
}
