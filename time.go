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
	if string(data) == "\"0001-01-01T00:00:00\"" {
		return nil
	}

	// Fractional seconds are handled implicitly by Parse.
	var err error
	t.Time, err = time.Parse(`"`+time.RFC3339+`"`, string(data))
	return err
}
