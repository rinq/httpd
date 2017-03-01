package statuspage

import (
	"bytes"
	"net/http"

	"github.com/golang/gddo/httputil/header"
)

// Write outputs an HTTP status page for code c to w, in response to r.
func Write(w http.ResponseWriter, r *http.Request, c int) (n int64, err error) {
	return WriteMessage(w, r, c, Message(c))
}

// WriteMessage outputs an HTTP status page for code c to w, in response to r,
// with a custom message m.
func WriteMessage(w http.ResponseWriter, r *http.Request, c int, m string) (int64, error) {
	var buf bytes.Buffer
	var contentType string
	context := context{c, http.StatusText(c), m}

	if useHTML(r) {
		if err := htmlTemplate.Execute(&buf, context); err == nil {
			contentType = "text/html"
		}
	}

	if contentType == "" {
		contentType = "text/plain"
		buf.Reset()
		textTemplate.Execute(&buf, context)
	}

	w.Header().Add("Content-Type", contentType+"; charset=utf-8")
	w.Header().Add("X-Status-Message", m)
	w.WriteHeader(c)
	return buf.WriteTo(w)
}

// context holds the data needed to render a status page.
type context struct {
	Code    int
	Text    string
	Message string
}

func useHTML(request *http.Request) bool {
	htmlQ := -1.0
	textQ := 0.0

	for _, spec := range header.ParseAccept(request.Header, "Accept") {
		if spec.Value == "text/html" || spec.Value == "application/xhtml+xml" {
			if spec.Q > htmlQ {
				htmlQ = spec.Q
			}
		} else if spec.Value == "text/plain" || spec.Value == "*.*" {
			if spec.Q > textQ {
				textQ = spec.Q
			}
		}
	}

	return htmlQ > textQ
}
