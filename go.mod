module github.com/artamonovkirill/random-photo

go 1.23.0

require (
	github.com/MaestroError/go-libheif v0.3.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.10.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// check it's available for RaspberryPI before updating
// https://pkgs.org/search/?q=libheif
require github.com/strukturag/libheif v1.17.6 // indirect
