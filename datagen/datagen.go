/*
This program is converts geojson files into go files containing structs that can
be read by rgeo. You can use this if you want to use a different dataset to any
of those included.

Usage

	go run datagen.go infile.geojson outfile.go

The variable containing the data will be named outfile. Currently rgeo will only
look for at the variable called countries110.
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

// Template for generated code
const tp = `// This file is generated

package rgeo

import geom "github.com/twpayne/go-geom"

var {{.Varname}} = rgeo{[]country{
	{{- range .Countries}}
	country{
		loc: Location{
			Country:	  "{{.Loc.Country}}",
			CountryLong:  "{{.Loc.CountryLong}}",
			CountryCode2: "{{.Loc.CountryCode2}}",
			CountryCode3: "{{.Loc.CountryCode3}}",
			Continent:	  "{{.Loc.Continent}}",
			Region:		  "{{.Loc.Region}}",
			SubRegion:	  "{{.Loc.SubRegion}}",
		},
		{{- if .Multi}}
		geo: newMPolyWithBounds(geom.{{.Layout}}, {{.Flatcoords}},
			{{.Bounds}}, {{.Ends}}),
		{{- else}}
		geo: newPolyWithBounds(geom.{{.Layout}}, {{.Flatcoords}},
			{{.Bounds}}, {{.Ends}}),
		{{- end}}
	},
	{{- end}}
}}
`

// viewData fills template tp
type viewData struct {
	Varname   string
	Countries []tpcountry
}

// tpcountry holds country data
type tpcountry struct {
	Loc rgeo.Location

	Multi      bool
	Layout     string
	Flatcoords string
	Ends       string
	Bounds     string
}

func main() {
	// Open infile
	infile, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	// Parse geojson
	var fc geojson.FeatureCollection
	if err := json.NewDecoder(infile).Decode(&fc); err != nil {
		panic(err)
	}

	var (
		countries   []tpcountry
		thisCountry tpcountry
	)

	for _, c := range fc.Features {
		thisCountry.Loc = getLocationStrings(c.Properties)

		g := c.Geometry
		switch g.(type) {
		case *geom.Polygon:
			thisCountry.Multi = false
			thisCountry.Ends = stringFromSlice(g.Ends())
		case *geom.MultiPolygon:
			thisCountry.Multi = true
			thisCountry.Ends = stringFromSlice(g.Endss())
		}

		thisCountry.Layout = fmt.Sprint(g.Layout())
		thisCountry.Flatcoords = stringFromSlice(g.FlatCoords())
		thisCountry.Bounds = stringFromSlice([]float64{g.Bounds().Min(0),
			g.Bounds().Min(1), g.Bounds().Max(0), g.Bounds().Max(1)})

		countries = append(countries, thisCountry)
	}

	infile.Close()

	// Open outfile
	outfile, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	w := bufio.NewWriter(outfile)

	// Create template
	tmpl, err := template.New("dat").Parse(tp)
	if err != nil {
		panic(err)
	}

	vd := viewData{strings.TrimSuffix(os.Args[2], ".go"), countries}

	// Write template
	err = tmpl.ExecuteTemplate(w, "dat", vd)
	if err != nil {
		panic(err)
	}

	w.Flush()
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
func getLocationStrings(p map[string]interface{}) rgeo.Location {
	country, ok := p["ADMIN"].(string)
	if !ok {
		country, ok = p["admin"].(string)
		if !ok {
			country = ""
		}
	}

	countrylong, ok := p["FORMAL_EN"].(string)
	if !ok {
		countrylong = ""
	}

	countrycode2, ok := p["ISO_A2"].(string)
	if !ok {
		countrycode2 = ""
	}

	countrycode3, ok := p["ISO_A3"].(string)
	if !ok {
		countrycode3 = ""
	}

	continent, ok := p["CONTINENT"].(string)
	if !ok {
		continent = ""
	}

	region, ok := p["REGION_UN"].(string)
	if !ok {
		region = ""
	}

	subregion, ok := p["SUBREGION"].(string)
	if !ok {
		subregion = ""
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
