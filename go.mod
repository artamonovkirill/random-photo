module github.com/artamonovkirill/random-photo

go 1.23.0

require (
	github.com/MaestroError/go-libheif v0.3.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
)

// check it's available for RaspberryPI before updating
// https://pkgs.org/search/?q=libheif
require github.com/strukturag/libheif v1.17.6 // indirect
