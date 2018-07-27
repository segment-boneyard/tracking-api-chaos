package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/segmentio/tracking-api-chaos/tracker"
)

func TestClient64(t *testing.T) {
	cases := []TTData{
		{
			name: "clientGroupBase64",
			req:  get("/v1/g", `{"groupId": "group-id"}`),
			headers: http.Header{
				"Content-Type":  {"application/json"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"groupId":"group-id","receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/g","headers":{}}`,
		},
		{
			name: "clientAliasBase64",
			req:  get("/v1/a", `{"userId": "user-id"}`),
			headers: http.Header{
				"Content-Type":  {"application/json"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"GET","path":"/v1/a","headers":{}}`,
		},
		{
			name: "clientPageBase64",
			req:  get("/v1/p", `{"name": "Docs"}`),
			headers: http.Header{
				"Content-Type":  {"application/json"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/p","headers":{}}`,
		},
		{
			name: "screenBase64",
			req:  get("/v1/s", `{"name": "Docs"}`),
			headers: http.Header{
				"Content-Type":  {"application/json"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/s","headers":{}}`,
		},
		{
			name: "clientTrackBase64",
			req:  get("/v1/t", `{"event": "Signup"}`),
			headers: http.Header{
				"Content-Type":  {"application/json"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/t","headers":{}}`,
		},
		{
			name: "clientBatchBase64",
			req:  get("/v1/b", `{"batch":[]}`),
			headers: http.Header{
				"Content-Type":  {"application/json"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/b","headers":{}}`,
		},
		{
			name: "clientJSONPBase64",
			req:  get("/v1/t?callback=log", `{"event": "Signup"}`),
			headers: http.Header{
				"Content-Type": {"text/javascript"},
			},
			code:     http.StatusOK,
			bodyResp: `typeof log == "function" && log({ success: true });`,
			nsqResp:  `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/t","headers":{}}`,
		},
		{
			name: "clientIdentifyBase64",
			req:  get("/v1/i", `{"userId": "user-id"}`),
			headers: http.Header{
				"Content-Type":  {"application/json"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"1970-01-01T00:00:50Z","userId":"user-id"},"method":"GET","path":"/v1/i","headers":{}}`,
			trackerNowFunc: func() time.Time {
				t := time.Unix(50, 0)
				return t.In(time.UTC)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, td := NewServerTest()
			defer td()

			if tc.trackerNowFunc != nil {
				oldTrackerFunc := tracker.Now
				tracker.Now = tc.trackerNowFunc
				defer func() { tracker.Now = oldTrackerFunc }()
			}

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
