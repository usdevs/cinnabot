package utils

import (
	t "time"
)

// SgLocation returns a Location representing Singapore's time zone (UTC+8).
func SgLocation() *t.Location {
	return t.FixedZone("+08", 8*60*60)
}
