package message

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

var tests = []struct {
	path   string
	expect bool
}{
	{"/v1/import", true},
	{"/v1/batch", true},
	{"/v1/b", true},
	{"/v1/track", false},
	{"/v1/identify", false},
}

func TestIsBatch(t *testing.T) {
	for _, test := range tests {
		m := &Message{Path: test.path}
		assert.Equal(t, m.Batch(), test.expect, "path: "+test.path)
	}
}

// TestClearProperty ensures that the different implementations of Payload support operations
// to clear top level properties
func TestClearProperty(t *testing.T) {
	const (
		keyFoo = "foo"
	)
	for _, test := range []struct {
		name        string
		payloadFunc func() Payload
	}{
		{
			name: "raw body type",
			payloadFunc: func() Payload {
				toRawMessage := func(in string) *json.RawMessage {
					msg := json.RawMessage([]byte(in))
					return &msg
				}
				body := make(RawBody)
				body[keyFoo] = toRawMessage(`"bar"`)
				return body
			},
		},
		{
			name: "body type",
			payloadFunc: func() Payload {
				body := make(Body)
				body[keyFoo] = "bar"
				return body
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			payload := test.payloadFunc()
			toMap := func(payload Payload) map[string]interface{} {
				b, err := json.Marshal(payload)
				if err != nil {
					t.Fatal(err)
				}
				var m map[string]interface{}
				if err := json.Unmarshal(b, &m); err != nil {
					t.Fatal(err)
				}
				return m
			}
			m := toMap(payload)
			if m[keyFoo] == nil {
				t.Fatal("foo should have been present")
			}
			m = toMap(payload)
			if m[keyFoo] == nil {
				t.Fatal("foo should have been present")
			}
		})
	}

}

func TestSetReceivedAt(t *testing.T) {
	{
		body := make(RawBody)
		time := time.Time{}
		err := body.SetReceivedAt(time)
		assert.Equal(t, err, nil)
		buf, err := json.Marshal(body)
		assert.Equal(t, err, nil)
		assert.Equal(t, string(buf), `{"receivedAt":"0001-01-01T00:00:00Z"}`)
	}
	{
		body := make(Body)
		time := time.Time{}
		err := body.SetReceivedAt(time)
		assert.Equal(t, err, nil)
		buf, err := json.Marshal(body)
		assert.Equal(t, err, nil)
		assert.Equal(t, string(buf), `{"receivedAt":"0001-01-01T00:00:00Z"}`)
	}
}

func TestFromQueryTrack(t *testing.T) {
	req := request("/v1/pixel/track?writeKey=foo&properties.name=baz&event=foo")
	msg, err := FromQuery("track", req)
	assert.Equal(t, err, nil)
	assert.Equal(t, msg.Path, "/v1/pixel/track")
	assert.Equal(t, msg.Method, "GET")
	assert.Equal(t, msg.Body, Body{
		"writeKey": "foo",
		"event":    "foo",
		"properties": Body{
			"name": "baz",
		},
	})
}

func TestFromQueryIdentify(t *testing.T) {
	req := request("/v1/pixel/track?writeKey=foo&traits.name=baz&traits.site=foo&userId=user")
	msg, err := FromQuery("track", req)
	assert.Equal(t, err, nil)
	assert.Equal(t, msg.Path, "/v1/pixel/track")
	assert.Equal(t, msg.Method, "GET")
	assert.Equal(t, msg.Body, Body{
		"writeKey": "foo",
		"userId":   "user",
		"traits": Body{
			"name": "baz",
			"site": "foo",
		},
	})
}

func TestFromQueryWithContext(t *testing.T) {
	req := request("/v1/pixel/track?writeKey=foo&traits.name=baz&userId=user&context.library.name=amp&context.library.version=1")
	msg, err := FromQuery("track", req)
	assert.Equal(t, err, nil)
	assert.Equal(t, msg.Path, "/v1/pixel/track")
	assert.Equal(t, msg.Method, "GET")
	assert.Equal(t, msg.Body, Body{
		"writeKey": "foo",
		"userId":   "user",
		"traits": Body{
			"name": "baz",
		},
		"context": Body{
			"library": Body{
				"name":    "amp",
				"version": "1",
			},
		},
	})
}

func TestEmptyKeys(t *testing.T) {
	q := make(url.Values)
	q.Add(" ", "")
	q.Add("", "")
	req := request("/v1/pixel/track?" + q.Encode())
	msg, err := FromQuery("track", req)
	assert.Equal(t, err, nil)
	assert.Equal(t, msg.Path, "/v1/pixel/track")
	assert.Equal(t, msg.Method, "GET")
	assert.Equal(t, msg.Body, Body{})
}

func TestQueryArrays(t *testing.T) {
	req := request("/v1/pixel/track?traits.user=a&traits.user=b")
	msg, err := FromQuery("track", req)
	assert.Equal(t, err, nil)
	assert.Equal(t, msg.Path, "/v1/pixel/track")
	assert.Equal(t, msg.Method, "GET")
	assert.Equal(t, msg.Body, Body{
		"traits": Body{
			"user": []string{"a", "b"},
		},
	})
}

func TestEmptyValues(t *testing.T) {
	req := request("/v1/pixel/track?a=&b=")
	msg, err := FromQuery("track", req)
	assert.Equal(t, err, nil)
	assert.Equal(t, msg.Path, "/v1/pixel/track")
	assert.Equal(t, msg.Method, "GET")
	assert.Equal(t, msg.Body, Body{"a": "", "b": ""})
}

func TestBadQuery(t *testing.T) {
	{ // should ignore `..foo`, since the first part is "empty"
		req := request("/v1/pixel/track?..foo=baz&event=event")
		msg, err := FromQuery("track", req)
		assert.Equal(t, err, nil)
		assert.Equal(t, msg.Path, "/v1/pixel/track")
		assert.Equal(t, msg.Method, "GET")
		assert.Equal(t, msg.Body, Body{"event": "event"})
	}
	{ // should ignore `foo. .`, since it contains empty parts
		req := request("/v1/pixel/track?foo. .=baz&event=event")
		msg, err := FromQuery("track", req)
		assert.Equal(t, err, nil)
		assert.Equal(t, msg.Path, "/v1/pixel/track")
		assert.Equal(t, msg.Method, "GET")
		assert.Equal(t, msg.Body, Body{"event": "event"})
	}
	{ // should not ignore `foo. b`
		req := request("/v1/pixel/track?foo. b=baz&event=event")
		msg, err := FromQuery("track", req)
		assert.Equal(t, err, nil)
		assert.Equal(t, msg.Path, "/v1/pixel/track")
		assert.Equal(t, msg.Method, "GET")
		assert.Equal(t, msg.Body, Body{"event": "event", "foo": Body{"b": "baz"}})
	}
}

func request(s string) *http.Request {
	req, _ := http.NewRequest("GET", "http://api.test"+s, nil)
	return req
}

func TestNew(t *testing.T) {
	req, _ := http.NewRequest("POST", "/v1/track", strings.NewReader(`{"writeKey":"1234567890"}`))
	req.Header.Set("Connection", "close") // hop-by-hop, doesn't get embedded in the message
	req.Header.Set("Content-Type", "application/json")

	msg := New(req)

	assert.Equal(t, msg.Body, nil)
	assert.Equal(t, msg.Method, "POST")
	assert.Equal(t, msg.Path, "/v1/track")
	assert.Equal(t, msg.Headers, http.Header{
		"Content-Type": {"application/json"},
	})
}
