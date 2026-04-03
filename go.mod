module github.com/zarishsphere/zs-core-fhir-r4-bridge

go 1.26.1

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/rs/zerolog v1.33.0
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	github.com/zarishsphere/zs-pkg-go-fhir v0.1.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
)

replace github.com/zarishsphere/zs-pkg-go-fhir => ../zs-pkg-go-fhir
