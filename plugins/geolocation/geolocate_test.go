package main

import "testing"

func TestGetStart(t *testing.T){
	t.Run("Test IPv6", func(t *testing.T){
		start := getStart("1a2f::1a2f1::1a2f")
		if start != "1a2f-" {
			t.Errorf("Expected 1a2f-, got %s", start)
		}
	})
	t.Run("Test IPv4", func(t *testing.T){
		start := getStart("1.0.1.1")
		if start != "1-0" {
			t.Errorf("Expected 1-0, got %s", start)
		}
	})
}
