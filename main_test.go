package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var location = time.Now().Location()

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func TestPrettifiesOrigins(t *testing.T) {
	for _, tc := range []struct {
		origin   string
		modified time.Time
		expected string
	}{
		{"/2013/2013 07 Somewhere/P1040926.JPG", date(2013, 07, 17), "Somewhere"},
		{"/2024/2024 02 Feb/IMG_4203.MOV", date(2024, 02, 17), ""},
		{"/2024/2024 07 July/IMG_4203.MOV", date(2024, 07, 17), ""},
		{"/2024/Somewhere/IMG_4203.MOV", date(2024, 02, 17), "Somewhere"},
	} {
		t.Run(fmt.Sprintf("%s->%s", tc.origin, tc.expected), func(t *testing.T) {
			assert.Equal(t, tc.expected, pretty(tc.origin, tc.modified))
		})
	}
}
