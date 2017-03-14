package native

import (
	"net"
	"net/http"

	"github.com/golang/gddo/httputil/header"
	"github.com/rinq/rinq-go/src/rinq"
)

// sessionAttributes returns the set of attributes to apply to new sessions for
// the given request.
func sessionAttributes(r *http.Request) []rinq.Attr {
	remoteAddr := ""
	for _, ip := range header.ParseList(r.Header, "X-Forwarded-For") {
		remoteAddr = ip
		break
	}

	if remoteAddr == "" {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if host != "" {
			remoteAddr = host
		} else {
			remoteAddr = r.RemoteAddr
		}
	}

	return []rinq.Attr{
		rinq.Freeze("rinq.httpd.remote-addr", remoteAddr),
		rinq.Freeze("rinq.httpd.host", r.Host),
	}
}
