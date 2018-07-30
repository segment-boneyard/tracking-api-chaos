package test

import (
	"testing"

	"net/http"
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
		srv := NewServerTest()
		srv.runTestCase(t, tc)
	}
}
