# rgeo ![](https://img.shields.io/github/workflow/status/sams96/rgeo/test?style=flat-square) [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/sams96/rgeo) [![](https://goreportcard.com/badge/github.com/sams96/rgeo?style=flat-square)](https://goreportcard.com/report/github.com/sams96/rgeo)

Package rgeo is a fast, simple solution for local reverse geocoding

Rather than relying on external software or online APIs, rgeo packages all of
the data it needs in your binary. This means it will only ever work down to the
level of cities (though currently just countries), but if that's all you need
then this is the library for you.

rgeo uses data from [naturalearthdata.com](https://naturalearthdata.com).

## Installation

    go get github.com/sams96/rgeo

## Usage

```go
loc, err := rgeo.ReverseGeocode([]float64{0, 52})
if err != nil {
	// Handle error
}

fmt.Printf("%s\n", loc.Country)
fmt.Printf("%s\n", loc.CountryLong)
fmt.Printf("%s\n", loc.CountryCode2)
fmt.Printf("%s\n", loc.CountryCode3)
fmt.Printf("%s\n", loc.Continent)
fmt.Printf("%s\n", loc.Region)
fmt.Printf("%s\n", loc.SubRegion)

// Output: United Kingdom
// United Kingdom of Great Britain and Northern Ireland
// GB
// GBR
// Europe
// Europe
// Northern Europe
```

Alternatively, if `ReverseGeocode` is being run multiple times, `New()` will
parse all of the data before hand so less work needs to be done each time.
```go
r, err := rgeo.New()
if err != nil {
	// Handle error
}

for i := -33; i <= 31; i += 5 {
	loc, err := r.ReverseGeocode([]float64{24, float64(i)})
	if err != nil {
		// Handle error
	}

	fmt.Printf("%s, ", loc.CountryCode2)
}

fmt.Printf("\n")

// Output: ZA, ZA, BW, NA, ZM, CD, CD, CD, CF, SD, SD, LY, LY,
```
## Contributing

Contributions are welcome, I haven't got any guidelines or anything so maybe
just make an issue first.
