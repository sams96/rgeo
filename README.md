# rgeo
[![](https://img.shields.io/github/workflow/status/sams96/rgeo/continuous-integration?style=for-the-badge)](https://github.com/sams96/rgeo/actions?query=workflow%3Acontinuous-integration)
[![](https://goreportcard.com/badge/github.com/sams96/rgeo?style=for-the-badge)](https://goreportcard.com/report/github.com/sams96/rgeo)
[![Release](https://img.shields.io/github/tag/sams96/rgeo.svg?label=release&color=24B898&logo=github&style=for-the-badge)](https://github.com/sams96/rgeo/releases/latest)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=for-the-badge)](https://pkg.go.dev/github.com/sams96/rgeo)

Package rgeo is a fast, simple solution for local reverse geocoding.

Rather than relying on external software or online APIs, rgeo packages all of
the data it needs in your binary. This means it will only ever work down to the
level of cities (though currently just countries), but if that's all you need
then this is the library for you.

rgeo uses data from [naturalearthdata.com](https://naturalearthdata.com), if
your coordinates are going to be near specific borders I would advise checking
the data beforehand (links to which are in the files).

## Installation

    go get github.com/sams96/rgeo

## Usage

Initialise rgeo using rgeo.New, which takes any number of datasets. The included
datasets are:
 - `Countries110` - Just country information, smallest and lowest detail of the
   included datasets.
 - `Countries10` - The same as above but with more detail.
 - `Provinces10` - Includes province information as well as country, so can
   still be used alone.
 - `Cities10` - Just city information, if you want provinces and/or countries as
   well use one of the above datasets with it.
Once initialised you can use `ReverseGeocode` on the value returned by `New`,
with your coordinates to get the location information. See the [Go
Docs](https://pkg.go.dev/github.com/sams96/rgeo) for more information on usage.

```go
r, err := rgeo.New(Countries110)
if err != nil {
	// Handle error
}

loc, err := r.ReverseGeocode([]float64{0, 52})
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

```go
r, err := New(Provinces10, Cities10)
if err != nil {
	// Handle error
}

loc, err := r.ReverseGeocode([]float64{141.35, 43.07})
if err != nil {
	// Handle error
}

fmt.Println(loc)
// Output: <Location> Sapporo, Hokkaido, Japan (JPN), Asia
```
## Contributing

Contributions are welcome, I haven't got any guidelines or anything so maybe
just make an issue first.

## Projects using rgeo

 - [rgeoSrv](https://github.com/sams96/rgeoSrv) - rgeo as a microservice
