package util

import (
	"net/url"
	"time"
)

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

// CleanPath returns the (decoded) path of the given URL, i.e., without query and fragment.
// An unparsable URL results in an empty path.
func CleanPath(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Path
}
