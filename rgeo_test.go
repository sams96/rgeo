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
	}{
		{
			name:     "Algeria",
			in:       []float64{1.880273, 31.787305},
			expected: "Algeria",
		},
		{
			name:     "Madagascar",
			in:       []float64{47.478275, -17.530126},
			expected: "Madagascar",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, err := ReverseGeocode(test.in)
			if err != nil {
				t.Logf("Error generated: %s\n", err)
				t.Fail()
			}
			if result != test.expected {
				t.Logf("expected: %s\ngot: %s\n", test.expected, result)
				t.Fail()
			}
		})
	}

}
