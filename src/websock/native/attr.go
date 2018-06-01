package native

import (
	"net"
	"net/http"

	"github.com/golang/gddo/httputil/header"
	"github.com/rinq/rinq-go/src/rinq"
)

const (
	//HttpdAttrNamespace is the namespace the attributes are in
	HttpdAttrNamespace = "rinq.httpd"
	//HttpdAttrHost contains the reported request host
	HttpdAttrHost = "host"
	//HttpdAttrClientIP contains the reported client host
	HttpdAttrClientIP = "client-ip"
	//HttpdAttrRemoteAddr contains the reported client host:port
	HttpdAttrRemoteAddr = "remote-addr"
	//HttpdAttrLocalAddr contains the report local host:port
	HttpdAttrLocalAddr = "local-addr"
)

// sessionAttributes returns the set of attributes to apply to new sessions for
// the given request.
func sessionAttributes(r *http.Request) []rinq.Attr {
	clientIP := ""
	for _, ip := range header.ParseList(r.Header, "X-Forwarded-For") {
		clientIP = ip
		break
	}

	if clientIP == "" {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if host != "" {
			clientIP = host
		} else {
			clientIP = r.RemoteAddr
		}
	}

	attr := []rinq.Attr{
		rinq.Freeze(HttpdAttrHost, r.Host),
		rinq.Freeze(HttpdAttrClientIP, clientIP),

		rinq.Freeze(HttpdAttrRemoteAddr, r.RemoteAddr),
	}

	if localAddr := r.Context().Value(http.LocalAddrContextKey); localAddr != nil {
		attr = append(attr, rinq.Freeze(HttpdAttrLocalAddr, localAddr.(net.Addr).String()))
	}

	return attr
}
