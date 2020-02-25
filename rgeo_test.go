/*
Copyright 2020 Sam Smith

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License.  You may obtain a copy of the
License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied.  See the License for the
specific language governing permissions and limitations under the License.
*/

package rgeo

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/go-test/deep"
	geom "github.com/twpayne/go-geom"
)

var testdata = []struct {
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
			Province:     "El Bayadh",
			ProvinceCode: "DZ-32",
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
			Province:     "Betsiboka",
			ProvinceCode: "MG-M",
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
			Province:     "Midlands",
			ProvinceCode: "ZW-MI",
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
			Province:     "Antarctica",
			ProvinceCode: "AQ-X01~",
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
			Province:     "Alaska",
			ProvinceCode: "US-AK",
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
			Province:     "Hertfordshire",
			ProvinceCode: "GB-HRT",
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
			Province:     "Al Kufrah",
			ProvinceCode: "LY-KF",
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
			Province:     "Al Wadi at Jadid",
			ProvinceCode: "EG-WAD",
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
			Province:     "North Dakota",
			ProvinceCode: "US-ND",
		},
	},
	{
		name: "Canada Border",
		in:   []float64{-102.560616, 49.02},
		err:  nil,
		expected: Location{
			Country:      "Canada",
			CountryLong:  "Canada",
			CountryCode2: "CA",
			CountryCode3: "CAN",
			Continent:    "North America",
			Region:       "Americas",
			SubRegion:    "Northern America",
			Province:     "Saskatchewan",
			ProvinceCode: "CA-SK",
		},
	},
}

func TestReverseGeocode_Countries(t *testing.T) {
	for _, dataset := range []func() []byte{Countries110, Countries10} {
		r, err := New(dataset)
		if err != nil {
			t.Error(err)
		}

		for _, test := range testdata {
			test := test

			test.expected.Province = ""
			test.expected.ProvinceCode = ""

			t.Run(test.name, func(t *testing.T) {
				result, err := r.ReverseGeocode(test.in)
				if err != test.err {
					t.Errorf("expected error: %s\n got: %s\n", test.err, err)
				}
				if diff := deep.Equal(test.expected, result); diff != nil {
					t.Error(diff)
				}
			})
		}
	}
}

func TestReverseGeocode_Provinces(t *testing.T) {
	for _, dataset := range []func() []byte{Provinces10} {
		r, err := New(dataset)
		if err != nil {
			t.Error(err)
		}

		for _, test := range testdata {
			test := test
			t.Run(test.name, func(t *testing.T) {
				result, err := r.ReverseGeocode(test.in)
				if err != test.err {
					t.Errorf("expected error: %s\n got: %s\n", test.err, err)
				}
				if diff := deep.Equal(test.expected, result); diff != nil {
					t.Error(diff)
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

func ExampleRgeo_ReverseGeocode() {
	r, err := New(Countries110)
	if err != nil {
		// Handle error
	}

	loc, err := r.ReverseGeocode([]float64{0, 52})
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
	r, err := New(Countries110)
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		})
	}
}

func BenchmarkReverseGeocode_10(b *testing.B) {
	r, err := New(Countries10)
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		})
	}
}

func BenchmarkReverseGeocode_Prov10(b *testing.B) {
	r, err := New(Provinces10)
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		})
	}
}
