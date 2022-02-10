package db

import (
	"testing"
	"time"
)

// Run some example durations and see how they look
func TestDurationFormat(t *testing.T) {
	compare(t, "730d", duration(2*365*24*time.Hour + 2*time.Minute))
	compare(t, "14d", duration(14*24*time.Hour + 30*time.Minute))
	compare(t, "2d", duration(48*time.Hour + 2*time.Minute))
	compare(t, "1d", duration(47*time.Hour + 48*time.Minute))
	compare(t, "2h", duration(2*time.Hour + 4*time.Minute))
	compare(t, "1h", duration(1*time.Hour + 59*time.Minute))
	compare(t, "30m", duration(30*time.Minute + 5*time.Second))
	compare(t, "2m", duration(2*time.Minute + 5*time.Second))
	compare(t, "5s", duration(5*time.Second))
}

func compare(t *testing.T, expect string, s string) {
	if expect != s {
		t.Errorf("Expected '%s' got '%s'", expect, s)
	}
}
