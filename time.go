package lenovo

import (
	"time"
)

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	// Fractional seconds are handled implicitly by Parse.
	var err error
	t.Time, err = time.Parse(`"`+time.RFC3339+`"`, string(data))
	if err != nil {
		// The eSupport API also returns timestamps without a timezone
		// offset (e.g. "2026-05-06T00:00:00"); interpret those as UTC.
		// This also covers the "0001-01-01T00:00:00" zero sentinel.
		t.Time, err = time.Parse(`"2006-01-02T15:04:05"`, string(data))
	}
	return err
}
