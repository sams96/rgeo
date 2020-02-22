/*
Package rgeo is a fast, simple solution for local reverse geocoding.

Rather than relying on external software or online APIs, rgeo packages all of
the data it needs in your binary. This means it will only ever work down to the
level of cities (though currently just countries), but if that's all you need
then this is the library for you.

rgeo uses data from https://naturalearthdata.com, if your coordinates are going
to be near specific borders I would advise checking the data beforehand (links
to which are in the files).

Installation

	go get github.com/sams96/rgeo

Contributing

Contributions are welcome, I haven't got any guidelines or anything so maybe
just make an issue first.
*/
package rgeo

import (
	"encoding/hex"
	"strings"

	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/xy"
)

// ErrLocationNotFound is returned when no country is found for given coordinates
var ErrLocationNotFound = errors.Errorf("country not found")

// Location is the return type for ReverseGeocode
type Location struct {
	// Commonly used country name
	Country string `json:"country,omitempty"`

	// Formal name of country
	CountryLong string `json:"country_long,omitempty"`

	// Two and three letter ISO 3166 codes
	CountryCode2 string `json:"country_code_2,omitempty"`
	CountryCode3 string `json:"country_code_3,omitempty"`

	Continent string `json:"continent,omitempty"`
	Region    string `json:"region,omitempty"`
	SubRegion string `json:"subregion,omitempty"`
}

// country hold the Polygon and Location for one country
type country struct {
	loc Location
	geo *s2.Polygon
}

// rgeo is the type used to hold pre-created polygons for reverse geocoding
type rgeo struct {
	countries []country
}

// ReverseGeocode returns the country in which the given coordinate is located
//
// The input is a geom.Coord, which is just a []float64 with the longitude
// in the zeroth position and the latitude in the first position.
// (i.e. []float64{lon, lat})
//
// The datasets currently available are Countries110, Countries50 & Countries10,
// these are returned from functions so the go compiler will ignore any that you
// aren't using. If using ReverseGeocode multiple times I would advise you to
// get the value from the function as a variable first and use that
func ReverseGeocode(loc geom.Coord, dataset *rgeo) (Location, error) {
	for _, feature := range dataset.countries {
		if in := feature.geo.ContainsPoint(pointFromCoord(loc)); in {
			return feature.loc, nil
		}
	}

	return Location{}, ErrLocationNotFound
}

// String method for type Location
func (l Location) String() string {
	// TODO: Add special case for empty Location
	ret := "<Location>"

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

// geometryContainsCoord checks if the given geometry (assuming that geometry is
// a polygon or multipolygon) contains the given point
func geometryContainsCoord(g geom.T, pt geom.Coord) bool {
	switch t := g.(type) {
	case *geom.Polygon:
		return polygonContainsCoord(t, pt)
	case *geom.MultiPolygon:
		return multiPolygonContainsCoord(t, pt)
	}

	return false
}

// polygonContainsCoord checks if the given coord is within the given polygon
func polygonContainsCoord(g *geom.Polygon, pt geom.Coord) bool {
	for i := 0; i < g.NumLinearRings(); i++ {
		r := g.LinearRing(i)
		if xy.IsPointInRing(r.Layout(), pt, r.FlatCoords()) {
			return true
		}
	}

	return false
}

// mutliPolygonContainsCoord checks if the given coord is within the given
// multipolygon
func multiPolygonContainsCoord(g *geom.MultiPolygon, pt geom.Coord) bool {
	for i := 0; i < g.NumPolygons(); i++ {
		r := g.Polygon(i)
		if polygonContainsCoord(r, pt) {
			return true
		}
	}

	return false
}

// From github.com/dgraph-io/dgraph
func pointFromCoord(r geom.Coord) s2.Point {
	// The geojson spec says that coordinates are specified as [long, lat]
	// We assume that any data encoded in the database follows that format.
	ll := s2.LatLngFromDegrees(r.Y(), r.X())
	return s2.PointFromLatLng(ll)
}

func decode(in string) *s2.Polygon {
	decode, _ := hex.DecodeString(in)
	p := new(s2.Polygon)
	p.Decode(strings.NewReader(string(decode)))
	return p
}
