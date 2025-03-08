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
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

// ErrLocationNotFound is returned when no country is found for given
// coordinates.
var ErrLocationNotFound = errors.New("country not found")

// Location is the return type for ReverseGeocode.
type Location struct {
	// Commonly used country name
	Country string `json:"country,omitempty"`

	// Formal name of country
	CountryLong string `json:"country_long,omitempty"`

	// ISO 3166-1 alpha-1 and alpha-2 codes
	CountryCode2 string `json:"country_code_2,omitempty"`
	CountryCode3 string `json:"country_code_3,omitempty"`

	Continent string `json:"continent,omitempty"`
	Region    string `json:"region,omitempty"`
	SubRegion string `json:"subregion,omitempty"`

	Province string `json:"province,omitempty"`

	// ISO 3166-2 code
	ProvinceCode string `json:"province_code,omitempty"`

	City string `json:"city,omitempty"`
}

// Rgeo is the type used to hold pre-created polygons for reverse geocoding.
type Rgeo struct {
	index *s2.ShapeIndex
	locs  map[s2.Shape]Location
}

// Go generate commands to regenerate the included datasets, this assumes you
// have the GeoJSON files from
// https://github.com/nvkelso/natural-earth-vector/tree/master/geojson.
// go run datagen/datagen.go -ne -o Countries110 ne_110m_admin_0_countries.geojson
// go run datagen/datagen.go -ne -o Countries10 ne_10m_admin_0_countries.geojson
// go run datagen/datagen.go -ne -o Provinces10 -merge ne_10m_admin_0_countries.geojson ne_10m_admin_1_states_provinces.geojson
// go run datagen/datagen.go -ne -o Cities10 ne_10m_urban_areas_landscan.geojson

// New returns an Rgeo struct which can then be used with ReverseGeocode. It
// takes any number of datasets as an argument. The included datasets are:
// Countries110, Countries10, Provinces10 and Cities10. Provinces10 includes all
// of the country information so if that's all you want don't use Countries as
// well. Cities10 only includes cities so you'll probably want to use
// Provinces10 with it.
func New(datasets ...func() []byte) (*Rgeo, error) {
	// Parse GeoJSON
	var fc geojson.FeatureCollection

	for i, dataset := range datasets {
		br := bytes.NewReader(dataset())
		if br.Len() == 0 {
			return nil, fmt.Errorf("no data in dataset %d", i)
		}

		zr, err := gzip.NewReader(br)
		if err != nil {
			return nil, fmt.Errorf("decompression failed for dataset %d: %w", i, err)
		}

		// Parse GeoJSON
		var tfc geojson.FeatureCollection
		if err := json.NewDecoder(zr).Decode(&tfc); err != nil {
			return nil, fmt.Errorf("invalid JSON in dataset %d: %w", i, err)
		}

		if err := zr.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip reader for dataset %d: %w", i, err)
		}

		fc.Features = append(fc.Features, tfc.Features...)
	}

	// Initialise Rgeo struct
	ret := new(Rgeo)
	ret.index = s2.NewShapeIndex()
	ret.locs = make(map[s2.Shape]Location)

	// Convert GeoJSON features from geom (multi)polygons to s2 polygons
	for _, c := range fc.Features {
		p, err := polygonFromGeometry(c.Geometry)
		if err != nil {
			return nil, fmt.Errorf("bad polygon in geometry: %w", err)
		}

		ret.index.Add(p)

		// The s2 ContainsPointQuery returns the shapes that contain the given
		// point, but I haven't found any way to attach the location information
		// to the shapes, so I use a map to get the information.
		ret.locs[p] = getLocationStrings(c.Properties)
	}

	return ret, nil
}

// Build builds the underlying shape index. This ensures that future calls to
// ReverseGeocode will be fast. If Build is not called, then the first lookup
// will build the index implicitly and experience a 1s+ delay.
func (r *Rgeo) Build() {
	r.index.Build()
}

// ReverseGeocode returns the country in which the given coordinate is located.
//
// The input is a geom.Coord, which is just a []float64 with the longitude
// in the zeroth position and the latitude in the first position
// (i.e. []float64{lon, lat}).
func (r *Rgeo) ReverseGeocode(loc geom.Coord) (Location, error) {
	query := s2.NewContainsPointQuery(r.index, s2.VertexModelOpen)
	res := query.ContainingShapes(pointFromCoord(loc))
	if len(res) == 0 {
		return Location{}, ErrLocationNotFound
	}

	return r.combineLocations(res), nil
}

// combineLocations combines the Locations for the given s2 Shapes.
func (r *Rgeo) combineLocations(s []s2.Shape) (l Location) {
	for _, shape := range s {
		loc := r.locs[shape]
		l = Location{
			Country:      firstNonEmpty(l.Country, loc.Country),
			CountryLong:  firstNonEmpty(l.CountryLong, loc.CountryLong),
			CountryCode2: firstNonEmpty(l.CountryCode2, loc.CountryCode2),
			CountryCode3: firstNonEmpty(l.CountryCode3, loc.CountryCode3),
			Continent:    firstNonEmpty(l.Continent, loc.Continent),
			Region:       firstNonEmpty(l.Region, loc.Region),
			SubRegion:    firstNonEmpty(l.SubRegion, loc.SubRegion),
			Province:     firstNonEmpty(l.Province, loc.Province),
			ProvinceCode: firstNonEmpty(l.ProvinceCode, loc.ProvinceCode),
			City:         firstNonEmpty(l.City, loc.City),
		}
	}

	return
}

// firstNonEmpty returns the first non empty parameter.
func firstNonEmpty(s ...string) string {
	for _, i := range s {
		if i != "" {
			return i
		}
	}

	return ""
}

// Get the relevant strings from the GeoJSON properties.
func getLocationStrings(p map[string]interface{}) Location {
	return Location{
		Country:      getPropertyString(p, "ADMIN", "admin"),
		CountryLong:  getPropertyString(p, "FORMAL_EN"),
		CountryCode2: getPropertyString(p, "ISO_A2"),
		CountryCode3: getPropertyString(p, "ISO_A3"),
		Continent:    getPropertyString(p, "CONTINENT"),
		Region:       getPropertyString(p, "REGION_UN"),
		SubRegion:    getPropertyString(p, "SUBREGION"),
		Province:     getPropertyString(p, "name"),
		ProvinceCode: getPropertyString(p, "iso_3166_2"),
		City:         strings.TrimSuffix(getPropertyString(p, "name_conve"), "2"),
	}
}

// getPropertyString gets the value from a map given the key as a string, or
// from the next given key if the previous fails.
func getPropertyString(m map[string]interface{}, keys ...string) (s string) {
	var ok bool
	for _, k := range keys {
		s, ok = m[k].(string)
		if ok {
			break
		}
	}

	return
}

// polygonFromGeometry converts a geom.T to an s2 Polygon.
func polygonFromGeometry(g geom.T) (*s2.Polygon, error) {
	var (
		polygon *s2.Polygon
		err     error
	)

	switch t := g.(type) {
	case *geom.Polygon:
		polygon, err = polygonFromPolygon(t)
	case *geom.MultiPolygon:
		polygon, err = polygonFromMultiPolygon(t)
	default:
		return nil, errors.New("needs Polygon or MultiPolygon")
	}

	if err != nil {
		return nil, err
	}

	return polygon, nil
}

// Converts a geom MultiPolygon to an s2 Polygon.
func polygonFromMultiPolygon(p *geom.MultiPolygon) (*s2.Polygon, error) {
	loops := make([]*s2.Loop, 0, p.NumPolygons())

	for i := 0; i < p.NumPolygons(); i++ {
		this, err := loopSliceFromPolygon(p.Polygon(i))
		if err != nil {
			return nil, err
		}

		loops = append(loops, this...)
	}

	return s2.PolygonFromLoops(loops), nil
}

// Converts a geom Polygon to an s2 Polygon.
func polygonFromPolygon(p *geom.Polygon) (*s2.Polygon, error) {
	loops, err := loopSliceFromPolygon(p)
	return s2.PolygonFromLoops(loops), err
}

// Converts a geom Polygon to slice of s2 Loop.
//
// Modified from types.loopFromPolygon from github.com/dgraph-io/dgraph.
func loopSliceFromPolygon(p *geom.Polygon) ([]*s2.Loop, error) {
	loops := make([]*s2.Loop, 0, p.NumLinearRings())

	for i := 0; i < p.NumLinearRings(); i++ {
		r := p.LinearRing(i)
		n := r.NumCoords()

		if n < 4 {
			return nil, errors.New("can't convert ring with less than 4 points")
		}

		if !r.Coord(0).Equal(geom.XY, r.Coord(n-1)) {
			return nil, fmt.Errorf(
				"last coordinate not same as first for polygon: %+v", p.FlatCoords())
		}

		// S2 specifies that the orientation of the polygons should be CCW.
		// However there is no restriction on the orientation in WKB (or
		// GeoJSON). To get the correct orientation we assume that the polygons
		// are always less than one hemisphere. If they are bigger, we flip the
		// orientation.
		reverse := isClockwise(r)
		l := loopFromRing(r, reverse)

		// Since our clockwise check was approximate, we check the cap and
		// reverse if needed.
		if l.CapBound().Radius().Degrees() > 90 {
			// Remaking the loop sometimes caused problems, this works better
			l.Invert()
		}

		loops = append(loops, l)
	}

	return loops, nil
}

// Checks if a ring is clockwise or counter-clockwise. Note: This uses the
// algorithm for planar polygons and doesn't work for spherical polygons that
// contain the poles or the antimeridan discontinuity. We use this as a fast
// approximation instead.
//
// From github.com/dgraph-io/dgraph
func isClockwise(r *geom.LinearRing) bool {
	// The algorithm is described here
	// https://en.wikipedia.org/wiki/Shoelace_formula
	var a float64

	n := r.NumCoords()

	for i := 0; i < n; i++ {
		p1 := r.Coord(i)
		p2 := r.Coord((i + 1) % n)
		a += (p2.X() - p1.X()) * (p1.Y() + p2.Y())
	}

	return a > 0
}

// From github.com/dgraph-io/dgraph
func loopFromRing(r *geom.LinearRing, reverse bool) *s2.Loop {
	// In WKB, the last coordinate is repeated for a ring to form a closed loop.
	// For s2 the points aren't allowed to repeat and the loop is assumed to be
	// closed, so we skip the last point.
	n := r.NumCoords()
	pts := make([]s2.Point, n-1)

	for i := 0; i < n-1; i++ {
		var c geom.Coord
		if reverse {
			c = r.Coord(n - 1 - i)
		} else {
			c = r.Coord(i)
		}

		pts[i] = pointFromCoord(c)
	}

	return s2.LoopFromPoints(pts)
}

// From github.com/dgraph-io/dgraph
func pointFromCoord(r geom.Coord) s2.Point {
	// The GeoJSON spec says that coordinates are specified as [long, lat]
	// We assume that any data encoded in the database follows that format.
	ll := s2.LatLngFromDegrees(r.Y(), r.X())
	return s2.PointFromLatLng(ll)
}

// String method for type Location.
func (l Location) String() string {
	ret := "<Location>"

	// Special case for empty location
	if l == (Location{}) {
		return ret + " Empty Location"
	}

	// Add city name
	if l.City != "" {
		ret += " " + l.City + ","
	}

	// Add province name
	if l.Province != "" {
		ret += " " + l.Province + ","
	}

	// Add country name
	if l.Country != "" {
		ret += " " + l.Country
	} else if l.CountryLong != "" {
		ret += " " + l.CountryLong
	}

	// Add country code in brackets
	if l.CountryCode3 != "" {
		ret += " (" + l.CountryCode3 + ")"
	} else if l.CountryCode2 != "" {
		ret += " (" + l.CountryCode2 + ")"
	}

	// Add continent/region
	if len(ret) > len("<Location>") {
		ret += ","
	}

	switch {
	case l.Continent != "":
		ret += " " + l.Continent
	case l.Region != "":
		ret += " " + l.Region
	case l.SubRegion != "":
		ret += " " + l.SubRegion
	}

	return ret
}
