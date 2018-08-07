package websock

import (
	"github.com/alecthomas/units"
	"golang.org/x/sync/semaphore"
	"time"
)

// Option is used to configure a given HTTPHandler via the constructor
type Option func(*httpHandler)

// LimitToOrigin configures the HTTPHandler to reject all requests
// whose Origin header doesn't conform to the given pattern
func LimitToOrigin(pattern string) Option {
	return func(h *httpHandler) {
		h.upgrader.CheckOrigin = newOriginChecker(pattern)
	}
}

// PingInterval configures the HTTPHandler to ping all connections
// periodically, waiting at least the given interval between pings
func PingInterval(interval time.Duration) Option {
	return func(h *httpHandler) {
		h.pingInterval = interval
	}
}

// MaxMessageSize configures the HTTPHandler to close connections
// that send a message larger than max
func MaxMessageSize(max units.MetricBytes) Option {
	return func(h *httpHandler) {
		h.maxIncomingMsgSize = max
	}
}

// LogTo configures the HTTPHandler to use the given logger
// for all internal logging
func LogTo(l Logger) Option {
	return func(h *httpHandler) {
		h.logger = l
	}
}

// MaxConcurrentCalls configures the HTTPHandler to limit the
// number of calls running concurrently to be perConn for each
// connection and globally for all currently running connections
func MaxConcurrentCalls(perConn, globally int) Option {
	return func(h *httpHandler) {
		h.globalLimit = semaphore.NewWeighted(int64(globally))
		h.maxCallsPerConn = int64(perConn)
	}
}

// DefaultHandler configures the HTTPHandler to use the given
// handle when no matching sub protocol can be found. If no
// default handle is given, it will return
func DefaultHandler(handle Handler) Option {
	return func(h *httpHandler) {
		h.defaultHandler = handle
	}
}
