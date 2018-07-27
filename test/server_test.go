package test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"errors"
	"fmt"
	"net/url"
	"time"

	"sync"

	"log"

	"github.com/bmizerany/assert"
	"github.com/segmentio/nsq-go"
	"github.com/segmentio/tracking-api-chaos/api"
	"github.com/segmentio/tracking-api-chaos/message"
	"github.com/segmentio/tracking-api-chaos/tracker"
)

type ServerTest struct {
	messages <-chan nsq.Message
	errorsC  <-chan nsq.Message
	timeout  time.Duration
	*api.Server
}

// TableTestData
type TTData struct {
	name           string
	req            *http.Request
	reqFunc        func() *http.Request
	code           int
	bodyResp       string
	nsqResp        string
	headers        http.Header
	trackerNowFunc func() time.Time
}

type serverOpt func(*ServerTest)

func waitOnProducer(p *nsq.Producer, wg *sync.WaitGroup) {
	for {
		if p.Connected() {
			wg.Done()
			return
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func buildProducer(topic string) *nsq.Producer {
	p, err := nsq.StartProducer(nsq.ProducerConfig{
		Address: "localhost:4150",
		Topic:   topic,
	})
	check(err)
	return p
}

func buildConsumer(topic string) *nsq.Consumer {
	c, err := nsq.StartConsumer(nsq.ConsumerConfig{
		Address: "localhost:4150",
		Topic:   topic,
		Channel: "test",
	})
	check(err)
	return c
}

func NewServerTest(opts ...serverOpt) (*ServerTest, func()) {
	st := &ServerTest{
		timeout: 1 * time.Second,
	}

	tracker.Now = func() time.Time { return time.Time{} }
	now := time.Now().Nanosecond()

	p := buildProducer(fmt.Sprintf("test-topic-%v", now))
	errorProducer := buildProducer(fmt.Sprintf("test-topic-errors-%v#ephemeral", now))

	// Now we need to await the producers connecting
	var wg = sync.WaitGroup{}
	wg.Add(1)
	go waitOnProducer(p, &wg)
	wg.Add(1)
	go waitOnProducer(errorProducer, &wg)
	wc := make(chan struct{})
	go func() {
		wg.Wait()
		wc <- struct{}{}
	}()

	timeout := 5 * time.Second
	log.Println("Wait for nsq producers to connect")
	select {
	case <-wc:
	case <-time.After(timeout):
		log.Println("Timed out waiting for nsq producers to connect")
		panic("nsq producers failed to connect")
	}

	c := buildConsumer(fmt.Sprintf("test-topic-%v", now))
	errorConsumer := buildConsumer(fmt.Sprintf("test-topic-errors-%v#ephemeral", now))

	apiAddr, _ := url.Parse("http://localhost:7777")

	st.messages = c.Messages()
	st.errorsC = errorConsumer.Messages()
	st.Server = api.New(p, errorProducer, apiAddr)

	// Apply any custom options
	for _, o := range opts {
		if o != nil {
			o(st)
		}
	}

	// Cleanup our consumers and producers
	return st, func() {
		p.Stop()
		errorProducer.Stop()
	}
}

func (st *ServerTest) consume() (nsq.Message, error) {
	select {
	case msg := <-st.messages:
		msg.Finish()
		return msg, nil
	case <-time.After(time.Second):
		return nsq.Message{}, errors.New("consume: timeout of 1s reached")
	}
}

func (st *ServerTest) consumeError() (nsq.Message, error) {
	select {
	case msg := <-st.errorsC:
		msg.Finish()
		return msg, nil
	case <-time.After(time.Second):
		return nsq.Message{}, errors.New("consume: timeout of 1s reached")
	}
}

func TestServer(t *testing.T) {
	cases := []TTData{
		{
			name:     "basic",
			req:      post("/v1/identify", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "gzip",
			req:      postGzip("/v1/identify", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "trailingSlash",
			req:      post("/v1/identify/", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "caseSensitivity",
			req:      post("/v1/IDENTIFY", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "caseSensitivitySlash",
			req:      post("/v1/IDENTIFY/", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/identify","headers":{}}`,
		},
		{
			name:     "group",
			req:      post("/v1/group", `{"groupId": "group-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"groupId":"group-id","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/group","headers":{}}`,
		},
		{
			name:     "groupGzip",
			req:      postGzip("/v1/group", `{"groupId": "group-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"groupId":"group-id","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/group","headers":{}}`,
		},
		{
			name:     "alias",
			req:      post("/v1/alias", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/alias","headers":{}}`,
		},
		{
			name:     "aliasGzip",
			req:      postGzip("/v1/alias", `{"userId": "user-id"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"receivedAt":"0001-01-01T00:00:00Z","userId":"user-id"},"method":"POST","path":"/v1/alias","headers":{}}`,
		},
		{
			name:     "page",
			req:      post("/v1/page", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/page","headers":{}}`,
		},
		{
			name:     "pageGzip",
			req:      postGzip("/v1/page", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/page","headers":{}}`,
		},
		{
			name:     "screen",
			req:      post("/v1/screen", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/screen","headers":{}}`,
		},
		{
			name:     "screenGzip",
			req:      postGzip("/v1/screen", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/screen","headers":{}}`,
		},
		{
			name:     "track",
			req:      post("/v1/track", `{"event": "Signup"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/track","headers":{}}`,
		},
		{
			name:     "trackGzip",
			req:      postGzip("/v1/track", `{"event": "Signup"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/track","headers":{}}`,
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
			nsqResp:  `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/track","headers":{"Authorization":["Basic d3JpdGUta2V5Og=="]}}`,
		},
		{
			name:     "batch",
			req:      post("/v1/batch", `{"batch":[]}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/batch","headers":{}}`,
		},
		{
			name:     "batchGzip",
			req:      postGzip("/v1/batch", `{"batch":[]}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/batch","headers":{}}`,
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
			nsqResp:  "/v1/batch\n:\n[message] error decoding json from request: json: cannot unmarshal array into Go value of type message.RawBody\n22\n[{\"a\": \"b\"},{\"c\":\"d\"}]",
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
			nsqResp:  "/v1/batch\nfoo:bar\n[message] error decoding json from request: json: cannot unmarshal array into Go value of type message.RawBody\n22\n[{\"a\": \"b\"},{\"c\":\"d\"}]",
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
			nsqResp:  "",
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
			nsqResp:  "/v1/batch\n:\n[message] error decoding json from request: json: cannot unmarshal array into Go value of type message.RawBody\n12\n[{\"a\": \"b\"}]",
		},
		{
			name:     "import",
			req:      post("/v1/import", `{"batch":[]}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/import","headers":{}}`,
		},
		{
			name:     "importGzip",
			req:      postGzip("/v1/import", `{"batch":[]}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/import","headers":{}}`,
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
			nsqResp:  `{"body":{"batch":[],"receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/import","headers":{}}`,
		},
		{
			name:     "badJson",
			req:      post("/v1/identify", `{"userId"`),
			code:     http.StatusBadRequest,
			bodyResp: "Bad Request\n",
			nsqResp:  "/v1/identify\n:\nunexpected EOF\n9\n{\"userId\"",
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
			nsqResp:  "",
		},
		{
			name:     "put",
			req:      put("/v1/page", `{"name": "Docs"}`),
			code:     http.StatusOK,
			bodyResp: `{"success":true}`,
			nsqResp:  `{"body":{"name":"Docs","receivedAt":"0001-01-01T00:00:00Z"},"method":"PUT","path":"/v1/page","headers":{}}`,
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

			var msg nsq.Message
			var err error
			if tc.code == http.StatusOK {
				msg, err = srv.consume()
			} else {
				msg, err = srv.consumeError()
			}

			if tc.nsqResp != "" {
				assert.Equal(t, err, nil)
			} else {
				assert.NotEqual(t, err, nil)
			}
			assert.Equal(t, string(msg.Body), tc.nsqResp)
		})
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
			nsqResp:  "/v1/identify\n:\n[message] error decoding json from request: http: request body too large\n65539",
		},
		{
			// TODO: largeJsonBatch never returns on consumeError, need to understand why.
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

			msg, err := srv.consumeError()

			if tc.nsqResp != "" {
				assert.Equal(t, err, nil)
				assert.Equal(t, true, strings.HasPrefix(string(msg.Body), tc.nsqResp))
			}
		})
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
			nsqResp:  `{"body":{"event":"Signup","receivedAt":"0001-01-01T00:00:00Z"},"method":"POST","path":"/v1/track","headers":{"Origin":["https://example.com"]}}`,
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
			if tc.bodyResp != "" {
				assert.Equal(t, rec.Body.String(), tc.bodyResp)
			}

			for k, v := range tc.headers {
				assert.Equal(t, v[0], rec.Header().Get(k))
			}

			msg, err := srv.consume()
			if tc.nsqResp != "" {
				assert.Equal(t, err, nil)
				assert.Equal(t, string(msg.Body), tc.nsqResp)
			}
		})
	}
}

func BenchmarkServerTrackSmall(b *testing.B) {
	buf := fixture("track.small.json")

	srv, td := NewServerTest()
	defer td()
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

	srv, td := NewServerTest()
	defer td()
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

	srv, td := NewServerTest()
	defer td()
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
