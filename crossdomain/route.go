package crossdomain

import (
	"net/http"
	"strconv"
)

// Crossdomain response.
var xml = []byte(`<cross-domain-policy><allow-http-request-headers-from domain="*.segment.io" headers="*"/><site-control permitted-cross-domain-policies="all"/><allow-access-from domain="*" secure="false"/></cross-domain-policy>`)

// Route serves crossdomain xml for adobe flash player.
func Route(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Length", strconv.Itoa(len(xml)))
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(200)
	w.Write(xml)
}
