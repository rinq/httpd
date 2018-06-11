package native

import "time"

// Option is used to configure the Handler
type Option interface {
	modify(*visitor)
}

// MaxSyncCallTimeout sets the maximum time a call can be
func MaxSyncCallTimeout(max time.Duration) Option {
	return &maxSyncCallTimeout{max}
}

type maxSyncCallTimeout struct {
	max time.Duration
}

func (m *maxSyncCallTimeout) modify(v *visitor) {
	v.syncCallTimeout = m.max
}
