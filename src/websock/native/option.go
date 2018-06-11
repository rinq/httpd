package native

import "time"

// Option is used to configure the Handler
type Option func(*visitor)

// MaxSyncCallTimeout sets the maximum time a call can be
func MaxSyncCallTimeout(max time.Duration) Option {
	return func(v *visitor) {
		v.syncCallTimeout = max
	}
}
