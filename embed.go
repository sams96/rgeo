package rgeo

// embedding files individually here to allow the linker to strip out unused ones
import _ "embed"

//go:embed data/Cities10.gz
var cities10 []byte

func Cities10() []byte {
	return cities10
}

//go:embed data/Countries10.gz
var countries10 []byte

func Countries10() []byte {
	return countries10
}

//go:embed data/Countries110.gz
var countries110 []byte

func Countries110() []byte {
	return countries110
}

//go:embed data/Provinces10.gz
var provinces10 []byte

func Provinces10() []byte {
	return provinces10
}
