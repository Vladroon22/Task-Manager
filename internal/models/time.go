package models

import (
	"fmt"
	"strings"
	"time"
)

type CustomDate time.Time

func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	// Try different formats
	formats := []string{
		"2006-01-02",
		time.DateTime,
		time.RFC3339,
	}
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			*cd = CustomDate(t)
			return nil
		}
	}
	return fmt.Errorf("unable to parse date: %s", s)
}

func (cd CustomDate) Time() time.Time {
	return time.Time(cd)
}
