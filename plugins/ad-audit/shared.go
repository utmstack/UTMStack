package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
)

func logStr(event *plugins.Event, key string) string {
	log := event.GetLog()
	if log == nil || log[key] == nil {
		return ""
	}
	return strings.TrimSpace(log[key].GetStringValue())
}

func logNumStr(event *plugins.Event, key string) string {
	log := event.GetLog()
	if log == nil || log[key] == nil {
		return ""
	}
	switch v := log[key].AsInterface().(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

func parseTime(s string) *time.Time {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

func laterTime(cur, next *time.Time) *time.Time {
	if next != nil && (cur == nil || next.After(*cur)) {
		return next
	}
	return cur
}

func firstSet(cur, next *time.Time) *time.Time {
	if cur != nil {
		return cur
	}
	return next
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
