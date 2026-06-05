package repository

import "time"

func parseTimestamp(s string) time.Time {
	for _, layout := range []string{
		"2006-01-02T15:04:05-07",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
