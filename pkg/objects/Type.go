package objects

import "time"

type Object struct {
	definition map[string]any
	changed    bool
	exists     bool
	created    time.Time
	updated    time.Time
}
