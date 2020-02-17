package rgeo

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
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
				Country:      "Algeria",
				CountryLong:  "People's Democratic Republic of Algeria",
				CountryCode2: "DZ",
				CountryCode3: "DZA",
				Continent:    "Africa",
				Region:       "Africa",
				SubRegion:    "Northern Africa",
			},
		},
		{
			name: "Madagascar",
			in:   []float64{47.478275, -17.530126},
			err:  nil,
			expected: Location{
				Country:      "Madagascar",
				CountryLong:  "Republic of Madagascar",
				CountryCode2: "MG",
				CountryCode3: "MDG",
				Continent:    "Africa",
				Region:       "Africa",
				SubRegion:    "Eastern Africa",
			},
		},
		{
			name: "Zimbabwe",
			in:   []float64{29.832875, -19.948725},
			err:  nil,
			expected: Location{
				Country:      "Zimbabwe",
				CountryLong:  "Republic of Zimbabwe",
				CountryCode2: "ZW",
				CountryCode3: "ZWE",
				Continent:    "Africa",
				Region:       "Africa",
				SubRegion:    "Eastern Africa",
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
				Country:      "Antarctica",
				CountryLong:  "",
				CountryCode2: "AQ",
				CountryCode3: "ATA",
				Continent:    "Antarctica",
				Region:       "Antarctica",
				SubRegion:    "Antarctica",
			},
		},
		{
			name: "Alaska",
			in:   []float64{-150.542, 66.3},
			err:  nil,
			expected: Location{
				Country:      "United States of America",
				CountryLong:  "United States of America",
				CountryCode2: "US",
				CountryCode3: "USA",
				Continent:    "North America",
				Region:       "Americas",
				SubRegion:    "Northern America",
			},
		},
		{
			name: "UK",
			in:   []float64{0, 52},
			err:  nil,
			expected: Location{
				Country:      "United Kingdom",
				CountryLong:  "United Kingdom of Great Britain and Northern Ireland",
				CountryCode2: "GB",
				CountryCode3: "GBR",
				Continent:    "Europe",
				Region:       "Europe",
				SubRegion:    "Northern Europe",
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
			if diff := deep.Equal(test.expected, result); diff != nil {
				t.Error(diff)
				t.Fail()
			}
		})
	}

}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		in       Location
		expected string
	}{
		{
			name: "Algeria",
			in: Location{
				Country:      "Algeria",
				CountryCode3: "DZA",
				Continent:    "Africa",
			},
			expected: "<Location> Algeria (DZA), Africa",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result := fmt.Sprintf("%s", test.in)
			if diff := deep.Equal(test.expected, result); diff != nil {
				t.Error(diff)
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
	fmt.Printf("%s\n", loc.CountryCode2)
	fmt.Printf("%s\n", loc.CountryCode3)
	fmt.Printf("%s\n", loc.Continent)
	fmt.Printf("%s\n", loc.Region)
	fmt.Printf("%s\n", loc.SubRegion)

	// Output: United Kingdom
	// United Kingdom of Great Britain and Northern Ireland
	// GB
	// GBR
	// Europe
	// Europe
	// Northern Europe
}
