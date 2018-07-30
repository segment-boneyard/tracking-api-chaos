package test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/segmentio/tracking-api-chaos/api"
	"github.com/segmentio/tracking-api-chaos/chaos"
	"github.com/segmentio/tracking-api-chaos/message"
	"github.com/segmentio/tracking-api-chaos/tracker"
)

type ServerTest struct {
	outbuf  *bytes.Buffer
	timeout time.Duration
	*api.Server
}

// TableTestData
type TTData struct {
	name string
	// request to make
	req *http.Request
	// func to generate request to make
	reqFunc func() *http.Request
	// expected HTTP response cod
	code int
	// expected response body
	bodyResp string
	// expected outMessage (not checked for ErrorOut)
	outMsg string
	// expected headers
	headers http.Header
	// func to patch tracker.Now with; for timing
	trackerNowFunc func() time.Time
}

func NewServerTest() *ServerTest {
	var outbuf bytes.Buffer

	return &ServerTest{
		outbuf:  &outbuf,
		Server:  api.New(&outbuf, chaos.NopChaos{}),
		timeout: 1 * time.Second,
	}
}

func (st *ServerTest) runTestCase(t *testing.T, tc TTData) {
	t.Run(tc.name, func(t *testing.T) {
		if tc.trackerNowFunc == nil {
			tc.trackerNowFunc = func() time.Time { return time.Time{} }
		}
		oldTrackerFunc := tracker.Now
		tracker.Now = tc.trackerNowFunc
		defer func() { tracker.Now = oldTrackerFunc }()

		rec := httptest.NewRecorder()
		req := tc.req
		if req == nil {
			req = tc.reqFunc()
		}
		st.ServeHTTP(rec, req)

		if tc.code != 0 {
			assert.Equal(t, rec.Code, tc.code)
		}
		if tc.bodyResp != "" {
			assert.Equal(t, rec.Body.String(), tc.bodyResp)
		}
		for k, v := range tc.headers {
			assert.Equal(t, v[0], rec.Header().Get(k))
		}

		if rec.Code == http.StatusOK && tc.outMsg != "" {
			var actualOutMsg []byte
			actualOutMsg, err := st.outbuf.ReadBytes('\n')
			if err != nil {
				t.Fatalf("failed to read line; got `%s`; err: %s", actualOutMsg, err)
			}
			actualOutMsg = actualOutMsg[:len(actualOutMsg)-1]
			assert.Equal(t, string(tc.outMsg), string(actualOutMsg))
		}
	})
}

func TestServer(t *testing.T) {
	cases := []TTData{
		{
			name:     "basic",
			req:      post("/v1/identify", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "gzip",
			req:      postGzip("/v1/identify", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "trailingSlash",
			req:      post("/v1/identify/", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "caseSensitivity",
			req:      post("/v1/IDENTIFY", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "caseSensitivitySlash",
			req:      post("/v1/IDENTIFY/", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "group",
			req:      post("/v1/group", `{"groupId": "group-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"groupId":"group-id","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/group","headers":{}}`,
		},
		{
			name:     "groupGzip",
			req:      postGzip("/v1/group", `{"groupId": "group-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"groupId":"group-id","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/group","headers":{}}`,
		},
		{
			name:     "alias",
			req:      post("/v1/alias", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/alias","headers":{}}`,
		},
		{
			name:     "aliasGzip",
			req:      postGzip("/v1/alias", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/alias","headers":{}}`,
		},
		{
			name:     "page",
			req:      post("/v1/page", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/page","headers":{}}`,
		},
		{
			name:     "pageGzip",
			req:      postGzip("/v1/page", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/page","headers":{}}`,
		},
		{
			name:     "screen",
			req:      post("/v1/screen", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/screen","headers":{}}`,
		},
		{
			name:     "screenGzip",
			req:      postGzip("/v1/screen", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/screen","headers":{}}`,
		},
		{
			name:     "track",
			req:      post("/v1/track", `{"event": "Signup"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/track","headers":{}}`,
		},
		{
			name:     "trackGzip",
			req:      postGzip("/v1/track", `{"event": "Signup"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/track","headers":{}}`,
		},
		{
			name: "trackBasicAuth",
			reqFunc: func() *http.Request {
				req := post("/v1/track", `{"event": "Signup"}`)
				req.SetBasicAuth("write-key", "")
				return req
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/track","headers":{"Authorization":["Basic d3JpdGUta2V5Og=="]}}`,
		},
		{
			name:     "batch",
			req:      post("/v1/batch", `{"batch":[]}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/batch","headers":{}}`,
		},
		{
			name:     "batchGzip",
			req:      postGzip("/v1/batch", `{"batch":[]}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/batch","headers":{}}`,
		},
		{
			name: "invalidMessageWithoutAuthHeader",
			reqFunc: func() *http.Request {
				r := bytes.NewReader([]byte(`[{"a": "b"},{"c":"d"}]`))
				req, err := http.NewRequest("POST", "http://api.test/v1/batch", r)
				check(err)
				return req
			},
			code:     http.StatusBadRequest,
			bodyResp: "Bad Request\n",
		},
		{
			name: "invalidMessageWithAuthHeader",
			reqFunc: func() *http.Request {
				r := bytes.NewReader([]byte(`[{"a": "b"},{"c":"d"}]`))
				req, err := http.NewRequest("POST", "http://api.test/v1/batch", r)
				check(err)
				req.SetBasicAuth("foo", "bar")
				return req
			},
			code:     http.StatusBadRequest,
			bodyResp: "Bad Request\n",
		},
		{
			name: "invalidGzipErrors",
			reqFunc: func() *http.Request {
				r := bytes.NewReader([]byte(`[{"a": "b"},{"c":"d"}]`))
				req, err := http.NewRequest("POST", "http://api.test/v1/batch", r)
				check(err)
				req.Header.Add("Content-Encoding", "gzip")
				return req
			},
			code:     http.StatusBadRequest,
			bodyResp: "{\"success\":false,\"message\":\"Malformed gzip content\"}",
			outMsg:   "",
		},
		{
			name: "validGzipButInvalidRequestErrors",
			reqFunc: func() *http.Request {
				var buf bytes.Buffer
				zw := gzip.NewWriter(&buf)
				zw.Write([]byte(`[{"a": "b"}]`))
				check(zw.Close())
				req, err := http.NewRequest("POST", "http://api.test/v1/batch", &buf)
				check(err)
				req.Header.Add("Content-Encoding", "gzip")
				return req
			},
			code:     http.StatusBadRequest,
			bodyResp: "Bad Request\n",
		},
		{
			name:     "import",
			req:      post("/v1/import", `{"batch":[]}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/import","headers":{}}`,
		},
		{
			name:     "importGzip",
			req:      postGzip("/v1/import", `{"batch":[]}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/import","headers":{}}`,
		},
		{
			name: "importGzipHeaderSpacing",
			reqFunc: func() *http.Request {
				req := postGzip("/v1/import", `{"batch":[]}`)
				req.Header.Set("Content-Encoding", " gzip ")
				return req
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/import","headers":{}}`,
		},
		{
			name:     "badJson",
			req:      post("/v1/identify", `{"userId"`),
			code:     http.StatusBadRequest,
			bodyResp: "Bad Request\n",
		},
		{
			name: "badGzipContent",
			reqFunc: func() *http.Request {
				req := post("/v1/identify", `woot`)
				req.Header.Set("Content-Encoding", "gzip")
				return req
			},
			code:     http.StatusBadRequest,
			bodyResp: `{"success":false,"message":"Malformed gzip content"}`,
		},
		{
			name:     "put",
			req:      put("/v1/page", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"PUT","path":"/v1/page","headers":{}}`,
		},
	}

	for _, tc := range cases {
		srv := NewServerTest()
		srv.runTestCase(t, tc)
	}
}

func TestServerLargePayloads(t *testing.T) {
	cases := []TTData{
		{
			name: "largeJsonIdentify",
			reqFunc: func() *http.Request {
				huge := make([]int, message.Single+1)
				buf, err := json.Marshal(huge)
				check(err)
				assert.Equal(t, int64(len(buf)) > message.Single, true)
				req := post("/v1/identify", string(buf))
				return req
			},
			code:     http.StatusBadRequest,
			bodyResp: "Bad Request\n",
		},
		{
			name: "largeJsonBatch",
			reqFunc: func() *http.Request {
				huge := make([]int, message.Batch+1)
				buf, err := json.Marshal(huge)
				check(err)
				assert.Equal(t, int64(len(buf)) > message.Batch, true)
				req := post("/v1/batch", string(buf))
				return req
			},
			code:     http.StatusBadRequest,
			bodyResp: "Bad Request\n",
		},
	}

	for _, tc := range cases {
		srv := NewServerTest()
		srv.runTestCase(t, tc)
	}
}

func TestServerTracks(t *testing.T) {
	cases := []TTData{
		{
			name: "corsTrack",
			reqFunc: func() *http.Request {
				req := post("/v1/track", `{"event":"Signup"}`)
				req.Header.Set("Origin", "https://example.com")
				return req
			},
			headers: http.Header{
				"Access-Control-Allow-Origin": {"https://example.com"},
				"Content-Type":                {"application/json"},
			},
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			outMsg:   `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/track","headers":{"Origin":["https://example.com"]}}`,
		},
		{
			name: "preflightTrack",
			reqFunc: func() *http.Request {
				req := options("/v1/track")
				req.Header.Set("Origin", "http://example.com")
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
				return req
			},
			headers: http.Header{
				"Access-Control-Allow-Origin":      {"http://example.com"},
				"Access-Control-Allow-Methods":     {"POST"},
				"Access-Control-Allow-Headers":     {"Authorization, Content-Type"},
				"Access-Control-Max-Age":           {"604800"},
				"Access-Control-Allow-Credentials": {"true"},
			},
			code: http.StatusOK,
		},
	}

	for _, tc := range cases {
		srv := NewServerTest()
		srv.runTestCase(t, tc)
	}
}

func BenchmarkServerTrackSmall(b *testing.B) {
	buf := fixture("track.small.json")

	srv := NewServerTest()
	s := httptest.NewServer(srv)
	defer s.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := http.Post(s.URL+"/v1/track", "application/json", bytes.NewBuffer(buf))
			check(err)

			_, err = io.Copy(ioutil.Discard, res.Body)
			check(err)

			res.Body.Close()
		}
	})
}

func BenchmarkServerTrackLarge(b *testing.B) {
	buf := fixture("track.large.json")

	srv := NewServerTest()
	s := httptest.NewServer(srv)
	defer s.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := http.Post(s.URL+"/v1/track", "application/json", bytes.NewBuffer(buf))
			check(err)

			_, err = io.Copy(ioutil.Discard, res.Body)
			check(err)

			res.Body.Close()
		}
	})
}

func BenchmarkServerTrackBroken(b *testing.B) {
	buf := fixture("track.broken.json")

	srv := NewServerTest()
	s := httptest.NewServer(srv)
	defer s.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := http.Post(s.URL+"/v1/track", "application/json", bytes.NewBuffer(buf))
			check(err)

			_, err = io.Copy(ioutil.Discard, res.Body)
			check(err)

			res.Body.Close()
		}
	})
}
