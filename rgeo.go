package rgeo

import (
	"context"
	"encoding/json"
	"golang.org/x/net/webdav"
	"os"

	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

var ErrCountryNotFound = errors.Errorf("Country not found")

const geoDataPath = "ne_110m_admin_0_countries.geojson"

// reverseGeocode returns the country in which the given coordinate is located
func ReverseGeocode(loc geom.Coord) (string, error) {
	var dir webdav.Dir
	geoData, err := dir.OpenFile(context.Background(), geoDataPath, 0, os.ModeExclusive)
	if err != nil {
		return "", err
	}
	defer geoData.Close()

	var fc geojson.FeatureCollection
	if err := json.NewDecoder(geoData).Decode(&fc); err != nil {
		return "", err
	}

	for _, country := range fc.Features {
		switch geo := country.Geometry.(type) {
		case *geom.Polygon:
			in, err := polygonContainsCoord(geo, loc)
			if err != nil {
				return "", err
			}
			if in {
				if name, ok := country.Properties["ADMIN"].(string); ok {
					return name, nil
				}

				return "", errors.Errorf("Name not found")
			}
		case *geom.MultiPolygon:
			for i := 0; i < geo.NumPolygons(); i++ {
				in, err := polygonContainsCoord(geo.Polygon(i), loc)
				if err != nil {
					return "", err
				}
				if in {
					if name, ok := country.Properties["ADMIN"].(string); ok {
						return name, nil
					}

					return "", errors.Errorf("Name not found")
				}
			}
		default:
			return "", errors.Errorf("Type not known")
		}
	}

	return "", ErrCountryNotFound
}

// polygonContainsCoord checks if a geom.Coord is within a *geom.Polygon
func polygonContainsCoord(geo *geom.Polygon, pt geom.Coord) (bool, error) {
	loop, err := loopFromPolygon(geo)
	if err != nil {
		return false, err
	}

	if loop.ContainsPoint(pointFromCoord(pt)) {
		return true, nil
	}

	return false, nil
}

// loopFromPolygon converts a geom.Polygon to a s2.Loop. We use loops instead of
// s2.Polygon as the s2.Polygon implementation is incomplete.
//
// Modified from github.com/dgraph-io/dgraph
func loopFromPolygon(p *geom.Polygon) (*s2.Loop, error) {
	// go implementation of s2 does not support more than one loop (and will
	// panic if the size of the loops array > 1). So we will skip the holes in
	// the polygon and just use the outer loop.
	r := p.LinearRing(0)
	n := r.NumCoords()
	if n < 4 {
		return nil, errors.Errorf("Can't convert ring with less than 4 pts")
	}
	if !r.Coord(0).Equal(geom.XY, r.Coord(n-1)) {
		return nil, errors.Errorf(
			"Last coordinate not same as first for polygon: %+v\n", p)
	}
	// S2 specifies that the orientation of the polygons should be CCW.  However
	// there is no restriction on the orientation in WKB (or geojson). To get
	// the correct orientation we assume that the polygons are always less than
	// one hemisphere. If they are bigger, we flip the orientation.
	reverse := isClockwise(r)
	l := loopFromRing(r, reverse)

	// Since our clockwise check was approximate, we check the cap and reverse
	// if needed.
	if l.CapBound().Radius().Degrees() > 90 {
		// Remaking the loop sometimes caused problems, this works better
		l.Invert()
	}
	return l, nil
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
