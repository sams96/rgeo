/*
Package rgeo is a fast, simple solution for local reverse geocoding

Rather than relying on external software or online APIs, rgeo packages all of
the data it needs in your binary. This means it will only ever work down to the
level of cities (though currently just countries), but if that's all you need
then this is the library for you.

rgeo uses data from https://naturalearthdata.com.

Installation

	go get github.com/sams96/rgeo

Contributing

Contributions are welcome, I haven't got any guidelines or anything so maybe
just make an issue first.
*/
package rgeo

import (
	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
	geom "github.com/twpayne/go-geom"
)

var errCountryNotFound = errors.Errorf("country not found")

// Location is the return type for ReverseGeocode
type Location struct {
	// Commonly used country name
	Country string

	// Formal name of country
	CountryLong string

	// Two and three letter ISO 3166 codes
	CountryCode2 string
	CountryCode3 string

	Continent string
	Region    string
	SubRegion string
}

// country hold the Polygon and Location for one country
type country struct {
	loc Location
	geo geom.T
}

// rgeo is the type used to hold pre-created polygons for reverse geocoding
type rgeo struct {
	countries []country
}

// ReverseGeocode returns the country in which the given coordinate is located
//
// The input is a `geom.Coord`, which is just a `[]float64` with the longitude
// in the zeroth position and the latitude in the first position.
// (i.e. `[]float64{lon, lat}`)
//
// When run without a type rgeo it re-creates the polygons every time
func ReverseGeocode(loc geom.Coord) (Location, error) {
	return countries110.ReverseGeocode(loc)
}

// ReverseGeocode returns the country in which the given coordinate is located
//
// The input is a `geom.Coord`, which is just a `[]float64` with the longitude
// in the zeroth position and the latitude in the first position.
// (i.e. `[]float64{lon, lat}`)
//
// When run on a type rgeo it uses the pre-created polygons instead of
// calculating them every time
func (r *rgeo) ReverseGeocode(loc geom.Coord) (Location, error) {
	for _, country := range r.countries {
		poly, err := polygonFromGeometry(country.geo)
		if err != nil {
			return Location{}, err
		}

		if in := polygonContainsCoord(poly, loc); in {
			return country.loc, nil
		}
	}

	return Location{}, errCountryNotFound
}

// String method for type `Location`
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

// polygonContainsCoord checks if the given coord is within the given polygon
func polygonContainsCoord(p *s2.Polygon, pt geom.Coord) bool {
	return p.ContainsPoint(pointFromCoord(pt))
}

// polygonFromGeometry converts a geom.T to an s2.Polygon
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
		return nil, errors.Errorf("needs geom.Polygon or geom.MultiPolygon")
	}

	if err != nil {
		return nil, err
	}

	return polygon, nil
}

// Converts a `*geom.MultiPolygon` to an `*s2.Polygon`
func polygonFromMultiPolygon(p *geom.MultiPolygon) (*s2.Polygon, error) {
	var loops []*s2.Loop

	for i := 0; i < p.NumPolygons(); i++ {
		this, err := loopSliceFromPolygon(p.Polygon(i))
		if err != nil {
			return nil, err
		}

		loops = append(loops, this...)
	}

	return s2.PolygonFromLoops(loops), nil
}

// Converts a `*geom.Polygon` to an `*s2.Polygon`
func polygonFromPolygon(p *geom.Polygon) (*s2.Polygon, error) {
	loops, err := loopSliceFromPolygon(p)
	return s2.PolygonFromLoops(loops), err
}

// Converts a `*geom.Polygon` to slice of `*s2.Loop`
//
// Modified from types.loopFromPolygon from github.com/dgraph-io/dgraph
func loopSliceFromPolygon(p *geom.Polygon) ([]*s2.Loop, error) {
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
	// The geojson spec says that coordinates are specified as [long, lat]
	// We assume that any data encoded in the database follows that format.
	ll := s2.LatLngFromDegrees(r.Y(), r.X())
	return s2.PointFromLatLng(ll)
}
