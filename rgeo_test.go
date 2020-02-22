package rgeo

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/go-test/deep"
	geom "github.com/twpayne/go-geom"
)

func TestReverseGeocode_Countries(t *testing.T) {
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
			err:      ErrLocationNotFound,
			expected: Location{},
		},
		{
			name:     "North Pole",
			in:       []float64{-135, 90},
			err:      ErrLocationNotFound,
			expected: Location{},
		},
		{
			name: "South Pole",
			in:   []float64{44.99, -89.99},
			err:  nil,
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
		{
			name: "Libya",
			in:   []float64{24.98, 25.86},
			err:  nil,
			expected: Location{
				Country:      "Libya",
				CountryLong:  "Libya",
				CountryCode2: "LY",
				CountryCode3: "LBY",
				Continent:    "Africa",
				Region:       "Africa",
				SubRegion:    "Northern Africa",
			},
		},
		{
			name: "Egypt",
			in:   []float64{25.005187, 25.855963},
			err:  nil,
			expected: Location{
				Country:      "Egypt",
				CountryLong:  "Arab Republic of Egypt",
				CountryCode2: "EG",
				CountryCode3: "EGY",
				Continent:    "Africa",
				Region:       "Africa",
				SubRegion:    "Northern Africa",
			},
		},
		{
			name: "US Border",
			in:   []float64{-102.560616, 48.992073},
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
			name: "Canada Border",
			in:   []float64{-102.560616, 49.0},
			err:  nil,
			expected: Location{
				Country:      "Canada",
				CountryLong:  "Canada",
				CountryCode2: "CA",
				CountryCode3: "CAN",
				Continent:    "North America",
				Region:       "Americas",
				SubRegion:    "Northern America",
			},
		},
	}

	datasets := []func() *rgeo{Countries110, Countries50, Countries10,
		Provinces10}

	for i, dataset := range datasets {

		for _, test := range tests {
			test := test
			t.Run(test.name, func(t *testing.T) {
				result, err := ReverseGeocode(test.in, dataset())
				if err != test.err {
					t.Errorf("expected error: %s\n got: %s\n", test.err, err)
				}
				if diff := deep.Equal(test.expected, result); diff != nil {
					t.Error("In dataset", i, diff)
				}
			})
		}
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
		{
			name: "Zimbabwe",
			in: Location{
				CountryLong:  "Republic of Zimbabwe",
				CountryCode2: "ZW",
				Region:       "Africa",
			},
			expected: "<Location> Republic of Zimbabwe (ZW), Africa",
		},
		{
			name: "Northern America",
			in: Location{
				SubRegion: "Northern America",
			},
			expected: "<Location> Northern America",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result := test.in.String()
			if diff := deep.Equal(test.expected, result); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func ExampleReverseGeocode() {
	dataset := Countries110()
	loc, err := ReverseGeocode([]float64{0, 52}, dataset)
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

func BenchmarkReverseGeocode_110(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		}, Countries110())
	}
}

func BenchmarkReverseGeocode_110Pre(b *testing.B) {
	dataset := Countries110()
	for i := 0; i < b.N; i++ {
		_, _ = ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		}, dataset)
	}
}

func BenchmarkReverseGeocode_50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		}, Countries50())
	}
}

func BenchmarkReverseGeocode_50Pre(b *testing.B) {
	dataset := Countries50()
	for i := 0; i < b.N; i++ {
		_, _ = ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		}, dataset)
	}
}

func BenchmarkReverseGeocode_10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		}, Countries10())
	}
}

func BenchmarkReverseGeocode_10Pre(b *testing.B) {
	dataset := Countries10()
	for i := 0; i < b.N; i++ {
		_, _ = ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		}, dataset)
	}
}

func BenchmarkReverseGeocode_Prov(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		}, Provinces10())
	}
}

func BenchmarkReverseGeocode_ProvPre(b *testing.B) {
	dataset := Provinces10()
	for i := 0; i < b.N; i++ {
		_, _ = ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		}, dataset)
	}
}
