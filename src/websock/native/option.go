package native

import "time"

type option func(*visitor)

func SyncCallTimeout(max time.Duration) option {
	return func(v *visitor) {
		v.syncCallTimeout = &max
	}
}