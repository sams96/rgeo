# rgeo
[![](https://img.shields.io/github/actions/workflow/status/sams96/rgeo/continuous-integration.yml?branch=master&style=for-the-badge)](https://github.com/sams96/rgeo/actions?query=workflow%3Acontinuous-integration)
[![](https://goreportcard.com/badge/github.com/sams96/rgeo?style=for-the-badge)](https://goreportcard.com/report/github.com/sams96/rgeo)
[![Codecov](https://img.shields.io/codecov/c/github/sams96/rgeo?logo=codecov&style=for-the-badge)](https://codecov.io/gh/sams96/rgeo)
[![Release](https://img.shields.io/github/tag/sams96/rgeo.svg?label=release&color=24B898&logo=github&style=for-the-badge)](https://github.com/sams96/rgeo/releases/latest)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=for-the-badge)](https://pkg.go.dev/github.com/sams96/rgeo)

Rgeo is a fast, simple solution for local reverse geocoding, Rather than relying
on external software or online APIs, rgeo packages all of the data it needs in
your binary. This means it will only ever work down to the level of cities , but
if that's all you need then this is the library for you.

Rgeo uses data from [naturalearthdata.com](https://naturalearthdata.com), if
your coordinates are going to be near specific borders I would advise checking
the data beforehand (links to which are in the files). If you want to use your
own dataset, check out
[datagen](https://github.com/sams96/rgeo/tree/master/datagen).

## Key Features

 - **Fast** - So I haven't _actually_ benchmarked other reverse geocoding tools
   but on my laptop rgeo can run at under 800ns/op.
 - **Local** - Rgeo doesn't require pinging some API, most of which either cost
   money to use or have severe rate limits.
 - **Lightweight** - The rgeo repo is 32MB, which is large for a Go package but
   compared to the 800GB needed for a full planet install of
   [Nominatim](https://nominatim.org/release-docs/latest/admin/Installation/#hardware)
   it's miniscule.

## Installation

Download with

	go get github.com/sams96/rgeo

and add

```go
import "github.com/sams96/rgeo"
```

to the top of your Go file to include it in your project.

## Usage

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
// Output: <Location> Sapporo, Hokkaidō, Japan (JPN), Asia
```

First initialise rgeo using `rgeo.New`,
```go
func New(datasets ...func() []byte) (*Rgeo, error)
```
which takes any non-zero number of datasets as arguments. The included datasets
are:
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

Then use `ReverseGeocode` to get the location information of the given coordinate.

```go
func (r *Rgeo) ReverseGeocode(loc geom.Coord) (Location, error)
```

The input is a [`geom.Coord`](https://github.com/twpayne/go-geom), which is just
a `[]float64` with the longitude in the zeroth position and the latitude in the
first position (i.e. `[]float64{lon, lat}`). `ReverseGeocode` returns a
`Location`, which looks like this:

```go
type Location struct {
	// Commonly used country name
	Country string `json:"country,omitempty"`

	// Formal name of country
	CountryLong string `json:"country_long,omitempty"`

	// ISO 3166-1 alpha-1 and alpha-2 codes
	CountryCode2 string `json:"country_code_2,omitempty"`
	CountryCode3 string `json:"country_code_3,omitempty"`

	Continent string `json:"continent,omitempty"`
	Region    string `json:"region,omitempty"`
	SubRegion string `json:"subregion,omitempty"`

	Province string `json:"province,omitempty"`

	// ISO 3166-2 code
	ProvinceCode string `json:"province_code,omitempty"`

	City string `json:"city,omitempty"`
}
```

So, to put it all together:

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

## Contributing

Contributions are welcome, I haven't got any guidelines or anything so maybe
just make an issue first.

## Projects using rgeo

 - [rgeoSrv](https://github.com/sams96/rgeoSrv) - rgeo as a microservice
