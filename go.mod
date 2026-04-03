module github.com/zarishsphere/zs-core-fhir-r4-bridge

go 1.26.1

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/rs/zerolog v1.33.0
	github.com/stretchr/testify v1.10.0
	github.com/zarishsphere/zs-pkg-go-fhir v0.1.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/zarishsphere/zs-pkg-go-fhir => ../zs-pkg-go-fhir
