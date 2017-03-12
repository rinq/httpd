package websock

import "time"

// Config holds options for WebSocket connections.
type Config struct {
	OriginPattern string
	PingInterval  time.Duration
}
