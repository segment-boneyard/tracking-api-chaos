package message

import "testing"

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		b64 string
		str string
	}{
		{
			b64: `eyJ3cml0ZUtleSI6ImZFOGJGYVhkRkdvUU9HbnBiOTdnUGxsQTlRS3M4M3pkIiwidXNlcklkIjoiWm05MVlXUkFjMlZuYldWdWRDNWpiMjAiLCJwcm9wZXJ0aWVzIjp7InNlbmRlciI6ImZvdWFkQHNlZ21lbnQuY29tIiwic3ViamVjdCI6InlvIn0sImV2ZW50IjoiRW1haWwgT3BlbmVkIn0`,
			str: `{"writeKey":"fE8bFaXdFGoQOGnpb97gPllA9QKs83zd","userId":"Zm91YWRAc2VnbWVudC5jb20","properties":{"sender":"fouad@segment.com","subject":"yo"},"event":"Email Opened"}`,
		},
		{
			b64: `eyJ3cml0ZUtleSI6ImZFOGJGYVhkRkdvUU9HbnBiOTdnUGxsQTlRS3M4M3pkIiwidXNlcklkIjoiWm05MVlXUkFjMlZuYldWdWRDNWpiMjAiLCJwcm9wZXJ0aWVzIjp7InNlbmRlciI6ImZvdWFkQHNlZ21lbnQuY29tIiwic3ViamVjdCI6InlvIn0sImV2ZW50IjoiRW1haWwgT3BlbmVkIn0=`,
			str: `{"writeKey":"fE8bFaXdFGoQOGnpb97gPllA9QKs83zd","userId":"Zm91YWRAc2VnbWVudC5jb20","properties":{"sender":"fouad@segment.com","subject":"yo"},"event":"Email Opened"}`,
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
