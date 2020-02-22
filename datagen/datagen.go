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
	"encoding/json"
	"flag"
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
			{{- if .Multi}}
			geo: geom.NewMultiPolygonFlat(geom.{{.Layout}}, {{.Flatcoords}}, {{.Ends}}),
			{{- else}}
			geo: geom.NewPolygonFlat(geom.{{.Layout}}, {{.Flatcoords}}, {{.Ends}}),
			{{- end}}
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

	Multi      bool
	Layout     string
	Flatcoords string
	Ends       string
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
