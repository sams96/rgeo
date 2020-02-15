# rgeo

Package rgeo is a fast, simple solution for local reverse geocoding

Rather than relying on external software or online APIs, rgeo packages all of
the data it needs in your binary. This means it will only ever work down to the
level of cities (though currently just countries), but if that's all you need
then this is the library for you.

## Installation

    go get github.com/sams96/rgeo

## Usage

```go
loc, err := ReverseGeocode([]float64{0, 52})
if err != nil {
	// Handle error
}

fmt.Printf("%s\n", loc.Country)
fmt.Printf("%s\n", loc.CountryLong)
fmt.Printf("%s\n", loc.CountryCode)

// Output: United Kingdom
// United Kingdom of Great Britain and Northern Ireland
// GBR
```

## Contributing

Contributions are welcome, I haven't got any guidelines or anything so maybe
just make an issue first.
