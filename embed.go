package rgeo

import (
	"embed"
	"fmt"
	"io"
)

//go:embed data/*
var dataDir embed.FS

func Cities10() []byte {
	return readEmbedFile("Cities10.gz")
}

func Countries10() []byte {
	return readEmbedFile("Countries10.gz")
}

func Countries110() []byte {
	return readEmbedFile("Countries110.gz")
}

func Provinces10() []byte {
	return readEmbedFile("Provinces10.gz")
}

func readEmbedFile(name string) []byte {
	f, err := dataDir.Open(fmt.Sprintf("data/%s", name))
	if err != nil {
		panic(fmt.Sprintf("internal embedded geojson open error %v", err))
	}

	b, _ := io.ReadAll(f)
	return b
}
