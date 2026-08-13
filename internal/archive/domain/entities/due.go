package entities

import "time"

// Due alert classification for documents with due_date (in-app MVP).
const (
	DueAlertUpcoming  = "upcoming"
	DueAlertExpired   = "expired"
	DueSoonWindowDays = 7
)

// ClassifyDueDate returns upcoming, expired, or empty relative to now (calendar day).
// upcoming: due_date in [today, today+DueSoonWindowDays]; expired: due_date < today.
func ClassifyDueDate(due *time.Time, now time.Time) string {
	if due == nil {
		return ""
	}
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	d := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, loc)
	if d.Before(today) {
		return DueAlertExpired
	}
	horizon := today.AddDate(0, 0, DueSoonWindowDays)
	if !d.After(horizon) {
		return DueAlertUpcoming
	}
	return ""
}
