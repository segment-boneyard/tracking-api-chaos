package message

import "testing"

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		b64 string
		str string
	}{
		{
			b64: `eyJ3cml0ZUtleSI6ImFzZGYiLCJ1c2VySWQiOiIiLCJwcm9wZXJ0aWVzIjp7InNlbmRlciI6ImZyaWVuZHNAc2VnbWVudC5jb20iLCJzdWJqZWN0IjoieW8ifSwiZXZlbnQiOiJFbWFpbCBPcGVuZWQifQ`,
			str: `{"writeKey":"asdf","userId":"","properties":{"sender":"friends@segment.com","subject":"yo"},"event":"Email Opened"}`,
		},
		{
			b64: `eyJ3cml0ZUtleSI6ImFzZGYiLCJ1c2VySWQiOiIiLCJwcm9wZXJ0aWVzIjp7InNlbmRlciI6ImZyaWVuZHNAc2VnbWVudC5jb20iLCJzdWJqZWN0IjoieW8ifSwiZXZlbnQiOiJFbWFpbCBPcGVuZWQifQ==`,
			str: `{"writeKey":"asdf","userId":"","properties":{"sender":"friends@segment.com","subject":"yo"},"event":"Email Opened"}`,
		},
	}

	for _, test := range tests {
		if b, e := decodeBase64(test.b64); e != nil {
			t.Errorf("%s: error: %s", test.b64, e)
		} else if s := string(b); s != test.str {
			t.Errorf("%s: invalid string: %s", test.b64, s)
		}
	}
}
