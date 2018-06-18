package native

import "time"

// Option modifies how a given Handler handles messages from Rinq connections that the Handler manages.
// Options are typically applied to the Handler by passing them to NewHandler().
type Option interface {
	modify(*visitor)
}

// MaxCallTimeout sets the maximum time a call can be. The given time will
// override any timeout set by the client where the clients' timeout is
// longer.
func MaxCallTimeout(max time.Duration) Option {
	return &maxCallTimeout{max}
}

type maxCallTimeout struct {
	max time.Duration
}

func (m *maxCallTimeout) modify(v *visitor) {
	v.syncCallTimeout = m.max
}
