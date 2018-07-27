package test

import (
	"net/http/httptest"
	"testing"

	"net/http"

	"github.com/bmizerany/assert"
)

func TestCrossDomain(t *testing.T) {
	cases := []TTData{
		{
			name: "crossDomain",
			req:  get("/crossdomain.xml", ``),
			headers: http.Header{
				"Content-Length": {"210"},
				"Content-Type":   {"application/xml; charset=utf-8"},
			},
			code:     http.StatusOK,
			bodyResp: `<cross-domain-policy><allow-http-request-headers-from domain="*.segment.io" headers="*"/><site-control permitted-cross-domain-policies="all"/><allow-access-from domain="*" secure="false"/></cross-domain-policy>`,
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
		})
	}
}
