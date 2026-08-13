package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClassifyDueDate(t *testing.T) {
	now := time.Date(2026, 8, 12, 15, 30, 0, 0, time.UTC)
	day := func(y int, m time.Month, d int) *time.Time {
		t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		return &t
	}

	tests := []struct {
		name string
		due  *time.Time
		want string
	}{
		{name: "nil", due: nil, want: ""},
		{name: "expired yesterday", due: day(2026, 8, 11), want: DueAlertExpired},
		{name: "upcoming today", due: day(2026, 8, 12), want: DueAlertUpcoming},
		{name: "upcoming in 7 days", due: day(2026, 8, 19), want: DueAlertUpcoming},
		{name: "outside window day 8", due: day(2026, 8, 20), want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ClassifyDueDate(tt.due, now))
		})
	}
}
