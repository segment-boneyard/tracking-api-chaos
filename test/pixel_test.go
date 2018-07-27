package test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"encoding/json"

	"github.com/bmizerany/assert"
	"github.com/segmentio/tracking-api-chaos/message"
	"github.com/segmentio/tracking-api-chaos/pixel"
)

func TestPixels(t *testing.T) {
	cases := []TTData{
		{
			name: "pixelIdentify",
			req:  get("/v1/pixel/identify", `{"userId": "user-id"}`),
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"GET","path":"/v1/pixel/identify","headers":{}}`,
		},
		{
			name: "pixelIdentifyQuery",
			req:  query("/v1/pixel/identify", "userId=user&traits.name=name"),
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"receivedAt":"0001-01-01T00:00:00Z","traits":{"name":"name"},"userId":"user"},"method":"GET","path":"/v1/pixel/identify","headers":{}}`,
		},
		{
			name: "pixelGroup",
			req:  get("/v1/pixel/group", `{"groupId": "group-id"}`),
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"groupId":"group-id","receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/pixel/group","headers":{}}`,
		},
		{
			name: "pixelQueryGroup",
			req:  query("/v1/pixel/group", "groupId=group&traits.name=name"),
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"groupId":"group","receivedAt":"0001-01-01T00:00:00Z","traits":{"name":"name"}},"method":"GET","path":"/v1/pixel/group","headers":{}}`,
		},
		{
			name: "pixelAlias",
			req:  get("/v1/pixel/alias", `{"userId": "user-id"}`),
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"GET","path":"/v1/pixel/alias","headers":{}}`,
		},
		{
			name: "pixelQueryAlias",
			req:  query("/v1/pixel/alias", "userId=user"),
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user"},"method":"GET","path":"/v1/pixel/alias","headers":{}}`,
		},
		{
			name: "pixelPage",
			req:  get("/v1/pixel/page", `{"name": "Docs"}`),
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/pixel/page","headers":{}}`,
		},
		{
			name: "pixelQueryPage",
			reqFunc: func() *http.Request {
				q := make(url.Values)
				q.Set("name", "Docs")
				q.Set("properties.path", "/docs")
				req := query("/v1/pixel/page", q.Encode())
				return req
			},
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"name":"Docs","properties":{"path":"/docs"},"receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/pixel/page","headers":{}}`,
		},
		{
			name:    "pixelScreen",
			req:     get("/v1/pixel/screen", `{"name": "Docs"}`),
			code:    http.StatusOK,
			nsqResp: `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/pixel/screen","headers":{}}`,
		},
		{
			name: "pixelQueryScreen",
			reqFunc: func() *http.Request {
				q := make(url.Values)
				q.Set("name", "Docs")
				q.Set("properties.screen", "docs")
				q.Set("properties.path", "/docs")
				req := query("/v1/pixel/screen", q.Encode())
				return req
			},
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"name":"Docs","properties":{"path":"/docs","screen":"docs"},"receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/pixel/screen","headers":{}}`,
		},
		{
			name: "pixelTrack",
			req:  get("/v1/pixel/track", `{"event": "Signup"}`),
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"GET","path":"/v1/pixel/track","headers":{}}`,
		},
		{
			name: "pixelQueryTrack",
			req:  query("/v1/pixel/track", "userId=user&event=event&properties.foo.baz=baz&properties.value=1"),
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"event":"event","properties":{"foo":{"baz":"baz"},"value":"1"},"receivedAt":"0001-01-01T00:00:00Z","userId":"user"},"method":"GET","path":"/v1/pixel/track","headers":{}}`,
		},
		{
			name: "pixelBadJson",
			req:  get("/v1/pixel/identify", `{"userId"`),
			code: http.StatusOK,
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
			assert.Equal(t, rec.Body.Bytes(), pixel.GIF)
			for k, v := range tc.headers {
				assert.Equal(t, v[0], rec.Header().Get(k))
			}

			if tc.nsqResp != "" {
				msg, err := srv.consume()
				assert.Equal(t, err, nil)
				assert.Equal(t, string(msg.Body), tc.nsqResp)
			}
		})
	}
}

func TestPixelsLargeJson(t *testing.T) {
	cases := []TTData{
		{
			name: "pixelLargeJSON",
			reqFunc: func() *http.Request {
				huge := make([]int, message.Single+1)
				buf, err := json.Marshal(huge)
				check(err)
				assert.Equal(t, int64(len(buf)) > message.Single, true)
				req := get("/v1/pixel/identify", string(buf))
				return req
			},
			headers: http.Header{
				"Content-Type":  {"image/gif"},
				"Cache-Control": {"no-cache, max-age=0"},
			},
			code:    http.StatusOK,
			nsqResp: `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"GET","path":"/v1/pixel/identify","headers":{}}`,
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
		})
	}
}
