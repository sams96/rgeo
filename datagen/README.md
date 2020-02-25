# datagen

Command datagen converts GeoJSON files into go files containing functions that
return the GeoJSON, it can also merge properties from one GeoJSON file into
another using the -merge flag. You can use this if you want to use a different
dataset to any of those included, although that might be somewhat awkward if
the properties in your GeoJSON file are different.

### Usage

    go run datagen.go -o outfile.go infile.geojson

The variable containing the data will be named `outfile`.
