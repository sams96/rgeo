# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2023-01-03

It's been a while since the last release, so all of the dependencies have been
updated, but this is mostly about the first user contribution.

## Changed
 - Moved to using Go embed for the data files, thanks to @benjojo (#18)
 - Updated to Go 1.19
 - Updated other dependencies

## [1.1.1] - 2020-03-06

I broke v1.1.0 when removing the large files from the git history so this will
replace it.

### Changed
 - Much smaller repo
 - Readme rewrite

## [1.1.0] - 2020-02-28

Mostly just bumping the version to see if pkg.go.dev will update now the
data files are much smaller.

### Changed
 - GeoJSON data is now compressed with gzip in go files, leading to much smaller
   file sizes without a noticeable performance hit when reading.

## [1.0.0] - 2020-02-27

This release is now broken due to some large files being removed from the git
history.

### Added
 - Province information to `Location`.
 - New dataset `Cities10` and added `City` field to `Location`.

### Changed
 - Fixed datagen, and moved template back into code.
 - `New` can now accept multiple datasets (which is only really useful when you
   want to use `Cities10` and get province/country information.
 - Switched to Apache License.
 - Better errors.
 - 100% test coverage.

## [0.0.5] - 2020-02-23

This release is now broken due to some large files being removed from the git
history.

### Added
 - JSON strings in `Location` for more idiomatic marshalling .
 - datagen can now read multiple inputs to one output.
 - 1:10m scale datasets for countries and provinces.
 - Datasets are returned from functions so the compiler doesn't include unused
   ones in builds.
 - Brought back the `New` function to initialise the data.

### Changes
 - Massive speed increase on queries.
 - Data initialisation is probably slower.

## [0.0.4] - 2020-02-20

### Changed
 - Minor Documentation changes, making a new release just to get rid of the
   giant `This file is generated` in the godoc.

## [0.0.3] - 2020-02-19

### Added
 - More robust Location printing.

### Changed
 - Data is now included as a struct in a go file instead of geojson.
 - Changed the algorithm to remove dependence on s2, so it doesn't need to
   convert between s2 and geom types. This is a lot faster than converting to s2
   types every time, but slower than pre-converted.

### Removed
 - Function New() and access to the rgeo data type, the data is already parsed
   so it doesn't need to be parsed when used.

## [0.0.2] - 2020-02-17

### Added

 - 2 letter country codes, continents, regions and subregions to output.
 - Type `Rgeo` and function `New` to parse the JSON and create the polygons
   ahead of time so it doesn't need to be done every time `ReverseGeocode` is
   run.

### Changed

 - Moved to using s2 Polygons instead of just s2 Loops.
 - Using github.com/go-test/deep for nicer printing in tests.

## [0.0.1] - 2020-02-15

### Added

Initial release
 - Exposes Function `ReverseGeocode` and type `Location`.
 - Just give `ReverseGeocode` a pair of coordinates and it will return a
   `Location` containing information about which country those coordinates are
   in.
