/*
Command datagen converts geojson files into go files containing structs that can
be read by rgeo. You can use this if you want to use a different dataset to any
of those included.

Usage

	go run datagen.go -o outfile.go infile.geojson

The variable containing the data will be named outfile. Currently rgeo will only
look for at the variable called countries110.
*/
package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
	"github.com/sams96/rgeo"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

// Template for generated code
const tp = `// This file is generated

package rgeo

// {{.Varname}} {{.Comment}}
func {{.Varname}}() *rgeo {
	return &rgeo{[]country{
		{{- range .Countries}}
		{
			loc: Location{
				Country:      "{{.Loc.Country}}",
				CountryLong:  "{{.Loc.CountryLong}}",
				CountryCode2: "{{.Loc.CountryCode2}}",
				CountryCode3: "{{.Loc.CountryCode3}}",
				Continent:    "{{.Loc.Continent}}",
				Region:       "{{.Loc.Region}}",
				SubRegion:    "{{.Loc.SubRegion}}",
			},
			geo: decode("{{.Geo}}"),
		},
		{{- end}}
	}}
}
`

// viewData fills template tp
type viewData struct {
	Varname   string
	Comment   string
	Countries []tpcountry
}

// tpcountry holds country data
type tpcountry struct {
	Loc rgeo.Location

	Geo string
}

func main() {
	// Read args
	outFileName := flag.String("o", "", "Path to output file")
	neCommentFlag := flag.Bool("ne", false, "Use Natural earth comment")
	mergeFileName := flag.String("merge", "", "File to get extra info from")

	flag.Parse()

	if *outFileName == "" {
		fmt.Println("Please specify an output file with -o")
		return
	}

	feats, err := readInputs(flag.Args(), *mergeFileName)
	if err != nil {
		panic(err)
	}

	// Open outfile
	outfile, err := os.Create(*outFileName)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	w := bufio.NewWriter(outfile)

	var pre string
	if *neCommentFlag {
		pre = "https://github.com/nvkelso/natural-earth-vector/blob/master/geojson/"
	}

	files := flag.Args()
	if *mergeFileName != "" {
		files = append(files, *mergeFileName)
	}

	vd := viewData{
		Varname:   strings.TrimSuffix(*outFileName, ".go"),
		Comment:   "uses data from " + printSlice(prefixSlice(pre, files)),
		Countries: feats,
	}

	// Create template
	tmpl, err := template.New("tmpl").Parse(tp)
	if err != nil {
		panic(err)
	}

	// Write template
	err = tmpl.ExecuteTemplate(w, "tmpl", vd)
	if err != nil {
		panic(err)
	}

	w.Flush()
}

func readInputs(in []string, mergeFileName string) ([]tpcountry, error) {
	var feats []tpcountry

	var mergeData *[]tpcountry
	if mergeFileName != "" {
		md, err := readInput(mergeFileName, false, nil)
		if err != nil {
			return []tpcountry{}, err
		}
		mergeData = &md
	}

	for _, f := range in {
		s, err := readInput(f, true, mergeData)
		if err != nil {
			return []tpcountry{}, err
		}

		feats = append(feats, s...)
	}

	return feats, nil
}

func readInput(f string, withGeo bool, mergeData *[]tpcountry) ([]tpcountry, error) {
	// Open infile
	infile, err := os.Open(f)
	if err != nil {
		return []tpcountry{}, err
	}

	defer infile.Close()

	// Parse geojson
	var fc geojson.FeatureCollection
	if err := json.NewDecoder(infile).Decode(&fc); err != nil {
		return []tpcountry{}, err
	}

	var (
		thisCountry tpcountry
		feats       []tpcountry
	)

	for _, c := range fc.Features {
		thisCountry.Loc = getLocationStrings(c.Properties, mergeData)

		if !withGeo {
			feats = append(feats, thisCountry)
			continue
		}

		p, err := polygonFromGeometry(c.Geometry)
		if err != nil {
			return []tpcountry{}, err
		}

		buf := new(bytes.Buffer)
		err = p.Encode(buf)
		if err != nil {
			return []tpcountry{}, err
		}

		thisCountry.Geo = hex.EncodeToString(buf.Bytes())

		feats = append(feats, thisCountry)
	}

	return feats, nil
}

// stringFromSlice creates a string to represent a slice in generated code
func stringFromSlice(i interface{}) string {
	return fmt.Sprintf("%T%s", i,
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.Join(strings.Fields(fmt.Sprint(i)), ", "),
				"[", "{"),
			"]", "}"),
	)
}

// Get the relevant strings from the geojson properties
func getLocationStrings(p map[string]interface{}, mergeData *[]tpcountry) rgeo.Location {
	country, ok := p["ADMIN"].(string)
	if !ok {
		country, ok = p["admin"].(string)
		if !ok {
			country = ""
		}
	}

	var md rgeo.Location
	if mergeData != nil {
		md = findByCountry(country, mergeData)
	}

	countrylong, ok := p["FORMAL_EN"].(string)
	if !ok {
		countrylong = md.CountryLong
	}

	countrycode2, ok := p["ISO_A2"].(string)
	if !ok {
		countrycode2 = md.CountryCode2
	}

	countrycode3, ok := p["ISO_A3"].(string)
	if !ok {
		countrycode3 = md.CountryCode3
	}

	continent, ok := p["CONTINENT"].(string)
	if !ok {
		continent = md.Continent
	}

	region, ok := p["REGION_UN"].(string)
	if !ok {
		region = md.Region
	}

	subregion, ok := p["SUBREGION"].(string)
	if !ok {
		subregion = md.SubRegion
	}

	return rgeo.Location{
		Country:      country,
		CountryLong:  countrylong,
		CountryCode2: countrycode2,
		CountryCode3: countrycode3,
		Continent:    continent,
		Region:       region,
		SubRegion:    subregion,
	}
}

func findByCountry(country string, mergeData *[]tpcountry) rgeo.Location {
	for _, f := range *mergeData {
		if f.Loc.Country == country {
			return f.Loc
		}
	}

	return rgeo.Location{}
}

// printSlice prints a slice of strings with commas and an ampersand if needed
func printSlice(in []string) string {
	n := len(in)
	switch n {
	case 0:
		return ""
	case 1:
		return in[0]
	case 2:
		return strings.Join(in, " & ")
	default:
		return printSlice([]string{strings.Join(in[:n-1], ", "), in[n-1]})
	}
}

// prefix slice adds a given prefix to a slice of strings
func prefixSlice(pre string, slice []string) (ret []string) {
	for _, i := range slice {
		ret = append(ret, pre+i)
	}

	return
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
