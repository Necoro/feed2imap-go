package util

import "time"

// StrContains searches for `needle` in `haystack` and returns `true` if found.
func StrContains(haystack []string, needle string) bool {
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
