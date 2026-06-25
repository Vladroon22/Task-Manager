package models

import (
	"fmt"
	"strings"
	"time"
)

// In your models package
type CustomDate time.Time

func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		return nil
	}

	// Try date-only format first
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		*cd = CustomDate(t)
		return nil
	}

	// Fallback to RFC3339
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		*cd = CustomDate(t)
		return nil
	}

	return fmt.Errorf("invalid date format: %s", s)
}

func (cd CustomDate) Time() time.Time {
	return time.Time(cd)
}
