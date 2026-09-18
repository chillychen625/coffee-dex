package storage

import (
	"database/sql"
	"go-coffee-log/models"
	"time"
)

// Time helpers shared across Postgres storage implementations.
// Timestamps are stored in TEXT columns as RFC3339; dates (roast_date) as
// "2006-01-02". The fallback layout in parseTime covers rows written by the
// pre-migration SQLite database.

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func formatDateOnly(t time.Time) string {
	return t.Format("2006-01-02")
}

func parseTime(s sql.NullString) time.Time {
	if !s.Valid || s.String == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s.String); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s.String); err == nil {
		return t
	}
	return time.Time{}
}

func parseDateOnly(s sql.NullString) *models.DateOnly {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s.String)
	if err != nil {
		return nil
	}
	d := models.DateOnly(t)
	return &d
}
