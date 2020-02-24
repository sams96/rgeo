/*
Command datagen converts geojson files into go files containing functions that
return the geojson, it can also merge properties from one geojson file into
another using the -merge flag. You can use this if you want to use a different
dataset to any of those included.

Usage

	go run datagen.go -o outfile.go infile.geojson

The variable containing the data will be named outfile.
*/
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/twpayne/go-geom/encoding/geojson"
)

const tp = `// This file is generated

package rgeo

// {{.Varname}} {{.Comment}}
func {{.Varname}}() []byte {
	return []byte(` + "`" + `{{.JSON}}` + "`" + `)
}
`

// viewData fills template tp
type viewData struct {
	Varname string
	Comment string
	JSON    string
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
		log.Fatal(err)
	}

	// Open outfile
	outfile, err := os.Create(*outFileName)
	if err != nil {
		log.Fatal(err)
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

	resp, err := json.Marshal(feats)
	if err != nil {
		log.Fatal(err)
	}

	vd := viewData{
		Varname: strings.TrimSuffix(*outFileName, ".go"),
		Comment: "uses data from " + printSlice(prefixSlice(pre, files)),

		// I know this looks ridiculous, but it replaces backticks (which will
		// break the string) with `+"`"+`, which breaks the string, adds a
		// backtick and then restarts it
		JSON: strings.ReplaceAll(string(resp), "`", "`"+` + "`+"`"+`" + `+"`"),
	}

	// Create template
	tmpl, err := template.New("tmpl").Parse(tp)
	if err != nil {
		log.Fatal(err)
	}

	// Write template
	err = tmpl.ExecuteTemplate(w, "tmpl", vd)
	if err != nil {
		log.Fatal(err)
	}

	w.Flush()
}

func readInputs(in []string, mergeFileName string) (*geojson.FeatureCollection, error) {
	fc := new(geojson.FeatureCollection)

	var mergeData *geojson.FeatureCollection

	if mergeFileName != "" {
		md, err := readInput(mergeFileName, nil)
		if err != nil {
			return nil, err
		}

		mergeData = md
	}

	for _, f := range in {
		s, err := readInput(f, mergeData)
		if err != nil {
			return nil, err
		}

		fc.Features = append(fc.Features, s.Features...)
	}

	return fc, nil
}

func readInput(f string, mergeData *geojson.FeatureCollection) (*geojson.FeatureCollection, error) {
	// Open infile
	infile, err := os.Open(f)
	if err != nil {
		return nil, err
	}

	defer infile.Close()

	// Parse geojson
	var fc geojson.FeatureCollection
	if err := json.NewDecoder(infile).Decode(&fc); err != nil {
		return nil, err
	}

	if mergeData == nil {
		return &fc, nil
	}

	for _, feat := range fc.Features {
		country, ok := feat.Properties["admin"].(string)
		if !ok {
			log.Println("Country name in wrong place")
			break
		}

		for _, md := range mergeData.Features {
			if md.Properties["ADMIN"] == country {
				for k, v := range md.Properties {
					feat.Properties[k] = v
				}
			}
		}
	}

	return &fc, nil
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
