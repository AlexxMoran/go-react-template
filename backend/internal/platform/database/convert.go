package database

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// TimeOrZero unwraps a nullable timestamp into a time.Time (zero if NULL).
func TimeOrZero(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// TimePtr unwraps a nullable timestamp into a *time.Time (nil if NULL).
func TimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

// Timestamptz wraps a time.Time as a non-null pgtype value.
func Timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// TimestamptzPtr wraps a *time.Time, producing a NULL value when nil.
func TimestamptzPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// Text wraps a string as a non-null pgtype.Text.
func Text(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// TextPtr wraps a *string, producing a NULL value when nil.
func TextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// Int8Ptr wraps a *int64, producing a NULL value when nil.
func Int8Ptr(n *int64) pgtype.Int8 {
	if n == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *n, Valid: true}
}
