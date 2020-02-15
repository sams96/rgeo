/*
	Package rgeo is a fast, simple solution for local reverse geocoding

	Rather than relying on external software or online APIs, rgeo packages all
	of the data it needs in your binary. This means it will only ever work down
	to the level of cities (though currently just countries), but if that's all
	you need then this is the library for you.

	Installation

	    go get github.com/sams96/rgeo

	Contributing

	Contributions are welcome, I haven't got any guidelines or anything so maybe
	just make an issue first.
*/
package rgeo

import (
	"encoding/json"
	"fmt"

	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

var errCountryNotFound = errors.Errorf("country not found")
var errCountryLongNotFound = errors.Errorf("country long name not found")

type Location struct {
	Country     string
	CountryLong string
	CountryCode string
}

// reverseGeocode returns the country in which the given coordinate is located
//
// The input is a `geom.Coord`, which is just a `[]float64` with the longitude
// in the zeroth position and the latitude in the first position.
// (i.e. `[]float64{lon, lat}`)
func ReverseGeocode(loc geom.Coord) (Location, error) {
	var fc geojson.FeatureCollection
	if err := json.Unmarshal([]byte(geodata), &fc); err != nil {
		return Location{}, err
	}

	for _, country := range fc.Features {
		switch geo := country.Geometry.(type) {
		case *geom.Polygon:
			in, err := polygonContainsCoord(geo, loc)
			if err != nil {
				return Location{}, err
			}
			if in {
				return getLocationStrings(country)
			}
		case *geom.MultiPolygon:
			for i := 0; i < geo.NumPolygons(); i++ {
				in, err := polygonContainsCoord(geo.Polygon(i), loc)
				if err != nil {
					return Location{}, err
				}
				if in {
					return getLocationStrings(country)
				}
			}
		default:
			return Location{}, errors.Errorf("type not known")
		}
	}

	return Location{}, errCountryNotFound
}

// Get the relevant strings from the geojson feature
func getLocationStrings(f *geojson.Feature) (Location, error) {
	p := f.Properties
	var err error
	country, ok := p["ADMIN"].(string)
	if !ok {
		country, ok = p["admin"].(string)
		if !ok {
			err = errors.Errorf("country name not found")
		}
	}

	countrylong, ok := p["FORMAL_EN"].(string)
	if !ok {
		err = errCountryLongNotFound
	}

	countrycode, ok := p["ISO_A3"].(string)
	if !ok {
		err = errors.Errorf("country code name not found")
	}

	return Location{
		Country:     country,
		CountryLong: countrylong,
		CountryCode: countrycode,
	}, err
}

// String method for type `Location`
func (l Location) String() string {
	return fmt.Sprintf("<Location> %s, %s, %s", l.Country, l.CountryLong,
		l.CountryCode)
}

// polygonContainsCoord checks if a geom.Coord is within a *geom.Polygon
func polygonContainsCoord(geo *geom.Polygon, pt geom.Coord) (bool, error) {
	polygon, err := polygonFromPolygon(geo)
	if err != nil {
		return false, err
	}

	if polygon.ContainsPoint(pointFromCoord(pt)) {
		return true, nil
	}

	return false, nil
}

// Converts a `*geom.Polygon` to an `*s2.Polygon`
//
// Modified from types.loopFromPolygon from github.com/dgraph-io/dgraph
func polygonFromPolygon(p *geom.Polygon) (*s2.Polygon, error) {
	var loops []*s2.Loop
	for i := 0; i < p.NumLinearRings(); i++ {
		r := p.LinearRing(i)
		n := r.NumCoords()
		if n < 4 {
			return nil, errors.Errorf("Can't convert ring with less than 4 pts")
		}
		if !r.Coord(0).Equal(geom.XY, r.Coord(n-1)) {
			return nil, errors.Errorf(
				"Last coordinate not same as first for polygon: %+v\n", p)
		}

		// S2 specifies that the orientation of the polygons should be CCW.
		// However there is no restriction on the orientation in WKB (or
		// geojson). To get the correct orientation we assume that the polygons
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

	return s2.PolygonFromLoops(loops), nil
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
		p := pointFromCoord(c)
		pts[i] = p
	}
	return s2.LoopFromPoints(pts)
}

// From github.com/dgraph-io/dgraph
func pointFromCoord(r geom.Coord) s2.Point {
	// The geojson spec says that coordinates are specified as [long, lat]
	// We assume that any data encoded in the database follows that format.
	ll := s2.LatLngFromDegrees(r.Y(), r.X())
	return s2.PointFromLatLng(ll)
}
