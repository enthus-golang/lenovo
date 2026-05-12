package lenovo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		json string
		want time.Time
	}{
		{"null", `null`, time.Time{}},
		{"zero sentinel", `"0001-01-01T00:00:00"`, time.Time{}},
		{"RFC3339 UTC", `"2024-06-15T12:30:00Z"`, time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)},
		{"RFC3339 offset", `"2024-06-15T14:30:00+02:00"`,
			time.Date(2024, 6, 15, 14, 30, 0, 0, time.FixedZone("", 2*60*60))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got Time
			if err := json.Unmarshal([]byte(tc.json), &got); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Time = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTimeUnmarshalJSONInvalid(t *testing.T) {
	var got Time
	if err := json.Unmarshal([]byte(`"not a date"`), &got); err == nil {
		t.Fatal("Unmarshal succeeded, want error")
	}
}
