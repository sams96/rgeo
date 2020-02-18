package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sams96/rgeo"
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

	_, err = fmt.Fprintf(w, "package rgeo\n\n")
	_, err = fmt.Fprintf(w, "import \"github.com/golang/geo/s2\"\n\n")
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

		_, err = fmt.Fprintf(w, "\t\tPoly: s2.PolygonFromLoops([]*s2.Loop{\n")
		if err != nil {
			panic(err)
		}

		poly, err := rgeo.PolygonFromGeometry(c.Geometry)
		if err != nil {
			panic(err)
		}

		for _, l := range poly.Loops() {
			_, err = fmt.Fprintf(w, "\t\t\ts2.LoopFromPoints([]s2.Point{\n")
			if err != nil {
				panic(err)
			}
			for _, v := range l.Vertices() {
				_, err = fmt.Fprintf(w, "\t\t\t\ts2.PointFromCoords(%f, %f, %f),\n",
					v.X, v.Y, v.Z)
				if err != nil {
					panic("wrong len vertex")
				}
			}
			_, err = fmt.Fprintf(w, "\t\t\t}),\n")
			if err != nil {
				panic(err)
			}

		}
		_, err = fmt.Fprintf(w, "\t\t}),\n")
		if err != nil {
			panic(err)
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
