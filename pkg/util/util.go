package util

import "time"

// TimeFormat formats the given time, where an empty time is formatted as "not set".
func TimeFormat(t time.Time) string {
	if t.IsZero() {
		return "not set"
	}

	return t.Format(time.ANSIC)
}

// Days returns the number of days in a duration. Fraction of days are discarded.
func Days(d time.Duration) int {
	return int(d.Hours() / 24)
}
