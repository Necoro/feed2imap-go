package util

import "time"

// Contains searches for `needle` in `haystack` and returns `true` if found.
func Contains[T comparable](haystack []T, needle T) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}

	return false
}

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
