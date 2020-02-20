# datagen

This program is converts geojson files into go files containing structs that can
be read by rgeo. You can use this if you want to use a different dataset to any
of those included.

### Usage

    go run datagen.go infile.geojson outfile.go

The variable containing the data will be named `outfile`. Currently rgeo will only
look for at the variable called `countries110`.
