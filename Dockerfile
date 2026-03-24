# ZarishSphere — Multi-stage Dockerfile (Go 1.26.1)
# Stage 1: Build
FROM golang:1.26.1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /bin/server ./cmd/server/

# Stage 2: Run (minimal image)
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/server /server

EXPOSE 8000
USER nonroot:nonroot
ENTRYPOINT ["/server"]
