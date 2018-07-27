package test

import (
	"net/http/httptest"
	"testing"

	"net/http"

	"github.com/bmizerany/assert"
)

func TestJSON(t *testing.T) {
	cases := []TTData{
		{
			name: "clientIdentify",
			req:  post("/v1/i", `{"userId": "user-id"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/i","headers":{}}`,
		},
		{
			name: "clientGroup",
			req:  post("/v1/g", `{"groupId": "group-id"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"groupId":"group-id","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/g","headers":{}}`,
		},
		{
			name: "clientAlias",
			req:  post("/v1/a", `{"userId": "user-id"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/a","headers":{}}`,
		},
		{
			name: "clientPage",
			req:  post("/v1/p", `{"name": "Docs"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/p","headers":{}}`,
		},
		{
			name: "clientScreen",
			req:  post("/v1/s", `{"name": "Docs"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/s","headers":{}}`,
		},
		{
			name: "clientTrack",
			req:  post("/v1/t", `{"event":"Signup"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/t","headers":{}}`,
		},
		{
			name: "clientBatch",
			req:  post("/v1/b", `{"batch":[]}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/b","headers":{}}`,
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
