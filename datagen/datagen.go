/*
This program is converts geojson files into go files containing structs that can
be read by rgeo
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

const tp = `// This file is generated
package rgeo

import geom "github.com/twpayne/go-geom"

var {{.Varname}} = rgeo{[]country{
	{{- range .Countries}}
	country{
		loc: Location{
			Country: "{{.Loc.Country}}",
			CountryLong: "{{.Loc.CountryLong}}",
			CountryCode2: "{{.Loc.CountryCode2}}",
			CountryCode3: "{{.Loc.CountryCode3}}",
			Continent: "{{.Loc.Continent}}",
			Region: "{{.Loc.Region}}",
			SubRegion: "{{.Loc.SubRegion}}",
		},
		{{- if .Multi}}
		geo: geom.NewMultiPolygonFlat(geom.{{.Layout}}, {{.Flatcoords}}, {{.Ends}}),
		{{- else}}
		geo: geom.NewPolygonFlat(geom.{{.Layout}}, {{.Flatcoords}}, {{.Ends}}),
		{{- end}}
	},
	{{- end}}
}}
`

type viewData struct {
	Varname   string
	Countries []tpcountry
}

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
		thisCountry.Loc = getLocationStrings(c.Properties)

		switch g := c.Geometry.(type) {
		case *geom.Polygon:
			thisCountry.Multi = false
			thisCountry.Ends = stringFromSlice(g.Ends())
		case *geom.MultiPolygon:
			thisCountry.Multi = true
			thisCountry.Ends = stringFromSlice(g.Endss())
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

	tmpl, err := template.New("dat").Parse(tp)
	if err != nil {
		panic(err)
	}

	vd := viewData{strings.TrimSuffix(os.Args[2], ".go"), countries}

	err = tmpl.ExecuteTemplate(w, "dat", vd)
	if err != nil {
		panic(err)
	}
	w.Flush()
}

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
