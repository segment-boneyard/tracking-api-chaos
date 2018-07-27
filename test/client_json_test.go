package test

import (
	"testing"

	"net/http"
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
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/i","headers":{}}`,
		},
		{
			name: "clientGroup",
			req:  post("/v1/g", `{"groupId": "group-id"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"groupId":"group-id","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/g","headers":{}}`,
		},
		{
			name: "clientAlias",
			req:  post("/v1/a", `{"userId": "user-id"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/a","headers":{}}`,
		},
		{
			name: "clientPage",
			req:  post("/v1/p", `{"name": "Docs"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/p","headers":{}}`,
		},
		{
			name: "clientScreen",
			req:  post("/v1/s", `{"name": "Docs"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/s","headers":{}}`,
		},
		{
			name: "clientTrack",
			req:  post("/v1/t", `{"event":"Signup"}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/t","headers":{}}`,
		},
		{
			name: "clientBatch",
			req:  post("/v1/b", `{"batch":[]}`),
			headers: http.Header{
				"Content-Type": {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/b","headers":{}}`,
		},
	}

	for _, tc := range cases {
		srv := NewServerTest()
		srv.runTestCase(t, tc)
	}
}
