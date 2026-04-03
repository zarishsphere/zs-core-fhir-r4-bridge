FROM golang:1.21-alpine AS builder
RUN apk add --no-cache ca-certificates git tzdata
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ARG VERSION=1.0.0
ARG BUILD_TIME
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s -X 'main.Version=${VERSION}'" -o /build/zs-fhir-r4-bridge ./cmd/...

FROM gcr.io/distroless/static-debian12:nonroot
USER nonroot:nonroot
COPY --from=builder --chown=nonroot:nonroot /build/zs-fhir-r4-bridge /app/zs-fhir-r4-bridge
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
EXPOSE 8085
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 CMD ["/app/zs-fhir-r4-bridge","healthcheck"]
ENTRYPOINT ["/app/zs-fhir-r4-bridge"]
LABEL org.opencontainers.image.title="zs-fhir-r4-bridge"
LABEL org.opencontainers.image.vendor="ZarishSphere"
LABEL org.opencontainers.image.licenses="Apache-2.0"
