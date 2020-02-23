# datagen

Command datagen converts geojson files into go files containing functions that
return the geojson, it can also merge properties from one geojson file into
another using the -merge flag. You can use this if you want to use a different
dataset to any of those included.

### Usage

    go run datagen.go -o outfile.go infile.geojson

The variable containing the data will be named `outfile`.
