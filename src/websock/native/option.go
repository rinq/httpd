package native

import (
	"golang.org/x/sync/semaphore"
	"time"
)

// Option modifies how a given Handler handles messages from Rinq connections that the Handler manages.
// Options are typically applied to the Handler by passing them to NewHandler().
type Option interface {
	modify(*Handler)
}

// MaxCallTimeout sets the maximum time a call can be. The given time will
// override any timeout set by the client where the clients' timeout is
// longer.
func MaxCallTimeout(max time.Duration) Option {
	if max < 0 {
		panic("expected the maximum call timeout to be 0 or greater")
	}
	return &maxCallTimeout{max}
}

type maxCallTimeout struct {
	max time.Duration
}

func (m *maxCallTimeout) modify(h *Handler) {
	h.visitorOpt = append(h.visitorOpt, m.setTimeout)
}

func (m *maxCallTimeout) setTimeout(v *visitor) {
	v.syncCallTimeout = m.max
}

// MaxConcurrentCalls sets the maximum number of calls that will be processed concurrently. Any calls that
// are received while the Handler is at capacity will not be processed until there is spare capacity.
// Time waiting for spare capacity counts as part of against the message timeout.
func MaxConcurrentCalls(max int) Option {
	if max < 0 {
		panic("expected the number of maximum concurrent calls to be 0 or greater")
	}

	return maxConcurrentCalls(max)
}

type maxConcurrentCalls int

func (m maxConcurrentCalls) modify(h *Handler) {
	cl := semaphore.NewWeighted(int64(m))
	h.visitorOpt = append(h.visitorOpt, func(v *visitor) {
		v.syncCallCap = cl
	})
}
