/*
This is an experimental attempt at rewriting the geojson as go code so it
doesn't need to be pased every time. The main issues encountered were large
executables and very slow builds, the latter of which I think is due to having
to run a lot of s2 code to create the s2 polygons (because I can't access all of
the information in the structs so it has to re build them every time). However
once built it runs faster than previous methods.

Yes it is very ugly code, but I wrote it at 2am.
*/
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sams96/rgeo"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

func main() {
	infile, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	var fc geojson.FeatureCollection
	if err := json.NewDecoder(infile).Decode(&fc); err != nil {
		panic(err)
	}

	defer infile.Close()

	outfile, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	w := bufio.NewWriter(outfile)

	// Write package header
	_, err = fmt.Fprintf(w, "package rgeo\n\n")
	_, err = fmt.Fprintf(w, "import geom \"github.com/twpayne/go-geom\"\n\n")
	_, err = fmt.Fprintf(w, "var geodata2 = Rgeo{[]Country{\n") // TODO change var name
	if err != nil {
		panic(err)
	}

	for _, c := range fc.Features {
		_, err = fmt.Fprintf(w, "\tCountry{\n")

		loc := rgeo.GetLocationStrings(c.Properties)
		_, err = fmt.Fprintf(w, "\t\tLoc: Location{\n")
		_, err = fmt.Fprintf(w, "\t\t\tCountry: \"%s\",\n", loc.Country)
		_, err = fmt.Fprintf(w, "\t\t\tCountryLong: \"%s\",\n", loc.CountryLong)
		_, err = fmt.Fprintf(w, "\t\t\tCountryCode2: \"%s\",\n", loc.CountryCode2)
		_, err = fmt.Fprintf(w, "\t\t\tCountryCode3: \"%s\",\n", loc.CountryCode3)
		_, err = fmt.Fprintf(w, "\t\t\tContinent: \"%s\",\n", loc.Continent)
		_, err = fmt.Fprintf(w, "\t\t\tRegion: \"%s\",\n", loc.Region)
		_, err = fmt.Fprintf(w, "\t\t\tSubRegion: \"%s\",\n", loc.SubRegion)
		_, err = fmt.Fprintf(w, "\t\t},\n")
		if err != nil {
			panic(err)
		}

		switch g := c.Geometry.(type) {
		case *geom.Polygon:
			_, err = fmt.Fprintf(w,
				"\t\tPoly: geom.NewPolygonFlat(geom.%s, []float64{%s}, []int{%s}),\n",
				g.Layout(),
				stringFromSlice(g.FlatCoords()),
				stringFromSlice(g.Ends()))
			if err != nil {
				panic(err)
			}
		case *geom.MultiPolygon:
			_, err = fmt.Fprintf(w,
				"\t\tPoly: geom.NewMultiPolygonFlat(geom.%s, []float64{%s}, [][]int%s),\n",
				g.Layout(),
				stringFromSlice(g.FlatCoords()),
				stringFromDSlice(g.Endss()))
			if err != nil {
				panic(err)
			}
		default:
			panic("What what")
		}

		_, err = fmt.Fprintf(w, "\t},\n")
		w.Flush()
	}

	_, err = fmt.Fprintf(w, "}}")
	if err != nil {
		panic(err)
	}

	w.Flush()
}

func stringFromSlice(i interface{}) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(i)), ", "), "[]")
}

func stringFromDSlice(i interface{}) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.Join(strings.Fields(fmt.Sprint(i)), ", "),
			"[", "{"),
		"]", "}")
}
