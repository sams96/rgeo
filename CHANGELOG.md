# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- More robust Location printing

## [0.0.2] - 2020-02-17

### Added

 - 2 letter country codes, continents, regions and subregions to output
 - Type `Rgeo` and function `New` to parse the JSON and create the polygons
   ahead of time so it doesn't need to be done every time `ReverseGeocode` is
   run

### Changed

 - Moved to using s2 Polygons instead of just s2 Loops
 - Using github.com/go-test/deep for nicer printing in tests

## [0.0.1] - 2020-02-15

### Added

Initial release
 - Exposes Function `ReverseGeocode` and type `Location`
 - Just give `ReverseGeocode` a pair of coordinates and it will return a
   `Location` containing information about which country those coordinates are
   in.
