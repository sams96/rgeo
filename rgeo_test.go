package rgeo

import (
	"fmt"
	"testing"

	geom "github.com/twpayne/go-geom"
)

func TestReverseGeocode(t *testing.T) {
	tests := []struct {
		name     string
		in       geom.Coord
		err      error
		expected Location
	}{
		{
			name: "Algeria",
			in:   []float64{1.880273, 31.787305},
			err:  nil,
			expected: Location{
				Country:     "Algeria",
				CountryLong: "People's Democratic Republic of Algeria",
				CountryCode: "DZA",
			},
		},
		{
			name: "Madagascar",
			in:   []float64{47.478275, -17.530126},
			err:  nil,
			expected: Location{
				Country:     "Madagascar",
				CountryLong: "Republic of Madagascar",
				CountryCode: "MDG",
			},
		},
		{
			name: "Zimbabwe",
			in:   []float64{29.832875, -19.948725},
			err:  nil,
			expected: Location{
				Country:     "Zimbabwe",
				CountryLong: "Republic of Zimbabwe",
				CountryCode: "ZWE",
			},
		},
		{
			name:     "Ocean",
			in:       []float64{0, 0},
			err:      errCountryNotFound,
			expected: Location{},
		},
		{
			name:     "North Pole",
			in:       []float64{-135, 90},
			err:      errCountryNotFound,
			expected: Location{},
		},
		{
			name: "South Pole",
			in:   []float64{45, -90},
			err:  errCountryLongNotFound,
			expected: Location{
				Country:     "Antarctica",
				CountryLong: "",
				CountryCode: "ATA",
			},
		},
		{
			name: "Alaska",
			in:   []float64{-150.542, 66.3},
			err:  nil,
			expected: Location{
				Country:     "United States of America",
				CountryLong: "United States of America",
				CountryCode: "USA",
			},
		},
		{
			name: "UK",
			in:   []float64{0, 52},
			err:  nil,
			expected: Location{
				Country:     "United Kingdom",
				CountryLong: "United Kingdom of Great Britain and Northern Ireland",
				CountryCode: "GBR",
			},
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

func ExampleReverseGeocode() {
	loc, err := ReverseGeocode([]float64{0, 52})
	if err != nil {
		// Handle error
	}

	fmt.Printf("%s\n", loc.Country)
	fmt.Printf("%s\n", loc.CountryLong)
	fmt.Printf("%s\n", loc.CountryCode)

	// Output: United Kingdom
	// United Kingdom of Great Britain and Northern Ireland
	// GBR
}
