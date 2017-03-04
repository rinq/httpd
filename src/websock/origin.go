package websock

import (
	"net/http"
	"net/url"
	"strings"
)

// newOriginChecker returns a function that returns true if r's Origin header
// matches the pattern p.
func newOriginChecker(p string) func(r *http.Request) bool {
	p = strings.ToLower(p)

	// fallback to Gorilla's default, which matches against the Host header
	if p == "" {
		return nil
	}

	// match "*"
	if p == originWildcard {
		return func(*http.Request) bool {
			return true
		}
	}

	// match "*.domain.tld"
	if strings.HasPrefix(p, originWildcard) {
		suffix := strings.TrimPrefix(p, originWildcard)
		return func(r *http.Request) bool {
			return strings.HasSuffix(getOrigin(r), suffix)
		}
	}

	// match "host.*"
	if strings.HasSuffix(p, originWildcard) {
		prefix := strings.TrimSuffix(p, originWildcard)
		return func(r *http.Request) bool {
			return strings.HasPrefix(getOrigin(r), prefix)
		}
	}

	// match "host.domain.tld" exactly
	return func(r *http.Request) bool {
		return getOrigin(r) == p
	}
}

const originWildcard = "*"

func getOrigin(r *http.Request) string {
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return ""
	}

	u, err := url.Parse(origin[0])
	if err != nil {
		return ""
	}

	return strings.ToLower(u.Host)
}
