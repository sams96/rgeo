# rgeo
[![](https://img.shields.io/github/workflow/status/sams96/rgeo/continuous-integration?style=for-the-badge)](https://github.com/sams96/rgeo/actions?query=workflow%3Acontinuous-integration)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=for-the-badge)](https://pkg.go.dev/github.com/sams96/rgeo)
[![](https://goreportcard.com/badge/github.com/sams96/rgeo?style=for-the-badge)](https://goreportcard.com/report/github.com/sams96/rgeo)
[![Release](https://img.shields.io/github/tag/sams96/rgeo.svg?label=release&color=24B898&logo=github&style=for-the-badge)](https://github.com/sams96/rgeo/releases/latest)

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

## Contributing

Contributions are welcome, I haven't got any guidelines or anything so maybe
just make an issue first.
