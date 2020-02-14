package rgeo

import (
	"testing"

	geom "github.com/twpayne/go-geom"
)

func TestReverseGeocode(t *testing.T) {
	tests := []struct {
		name     string
		in       geom.Coord
		expected string
		err      error
	}{
		{
			name:     "Algeria",
			in:       []float64{1.880273, 31.787305},
			expected: "Algeria",
			err:      nil,
		},
		{
			name:     "Madagascar",
			in:       []float64{47.478275, -17.530126},
			expected: "Madagascar",
			err:      nil,
		},
		{
			name:     "Zimbabwe",
			in:       []float64{29.832875, -19.948725},
			expected: "Zimbabwe",
			err:      nil,
		},
		{
			name:     "Ocean",
			in:       []float64{0, 0},
			expected: "",
			err:      ErrCountryNotFound,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, err := ReverseGeocode(test.in)
			if err != test.err {
				t.Logf("expected error: %s\n got: %s\n", test.err, err)
				t.Fail()
			}
			if result != test.expected {
				t.Logf("expected: %s\ngot: %s\n", test.expected, result)
				t.Fail()
			}
		})
	}

}
