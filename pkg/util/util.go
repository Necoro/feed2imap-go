package util

import "time"

func StrContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}

	return false
}

func TimeFormat(t time.Time) string {
	if t.IsZero() {
		return "not set"
	} else {
		return t.Format(time.ANSIC)
	}
}
