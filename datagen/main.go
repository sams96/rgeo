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
	"text/template"

	"github.com/sams96/rgeo"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

const tp = `
// This file is generated
package rgeo

import geom "github.com/twpayne/go-geom"

var geodata2 = Rgeo{[]Country{
{{range .}}	Country{
		Loc: Location{
			Country: "{{.Loc.Country}}",
			CountryLong: "{{.Loc.CountryLong}}",
			CountryCode2: "{{.Loc.CountryCode2}}",
			CountryCode3: "{{.Loc.CountryCode3}}",
			Continent: "{{.Loc.Continent}}",
			Region: "{{.Loc.Region}}",
			SubRegion: "{{.Loc.SubRegion}}",
		},{{if .Multi}}
		Poly: geom.NewMultiPolygonFlat(geom.{{.Layout}}, {{.Flatcoords}}, {{.Ends}}),{{else}}
		Poly: geom.NewPolygonFlat(geom.{{.Layout}}, {{.Flatcoords}}, {{.Ends}}),{{end}}
	},
{{end}}}}
`

type tpcountry struct {
	Loc rgeo.Location

	Multi      bool
	Layout     string
	Flatcoords string
	Ends       string
}

func main() {
	infile, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	var fc geojson.FeatureCollection
	if err := json.NewDecoder(infile).Decode(&fc); err != nil {
		panic(err)
	}

	var (
		countries   []tpcountry
		thisCountry tpcountry
	)

	for _, c := range fc.Features {
		thisCountry.Loc = rgeo.GetLocationStrings(c.Properties)

		switch g := c.Geometry.(type) {
		case *geom.Polygon:
			thisCountry.Multi = false
			thisCountry.Ends = stringFromSlice(g.Ends())
		case *geom.MultiPolygon:
			thisCountry.Multi = true
			thisCountry.Ends = stringFromDSlice(g.Endss())
		}

		thisCountry.Layout = fmt.Sprint(c.Geometry.Layout())
		thisCountry.Flatcoords = stringFromSlice(c.Geometry.FlatCoords())

		countries = append(countries, thisCountry)
	}

	infile.Close()

	outfile, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	w := bufio.NewWriter(outfile)
	_ = w

	tmpl, err := template.New("dat").Parse(tp)
	if err != nil {
		panic(err)
	}
	err = tmpl.ExecuteTemplate(w, "dat", countries)
	if err != nil {
		panic(err)
	}
	w.Flush()
}

func stringFromSlice(i interface{}) string {
	return fmt.Sprintf("%T{%s}", i,
		strings.Trim(strings.Join(strings.Fields(fmt.Sprint(i)), ", "), "[]"))
}

func stringFromDSlice(i interface{}) string {
	return fmt.Sprintf("%T%s", i,
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.Join(strings.Fields(fmt.Sprint(i)), ", "),
				"[", "{"),
			"]", "}"),
	)
}
