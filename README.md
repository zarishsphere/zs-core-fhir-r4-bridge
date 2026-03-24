# zs-core-fhir-r4-bridge

> **Part of:** [ZarishSphere Platform](https://github.com/zarishsphere) | **Layer:** Backend Microservice
> **Language:** Go 1.26.1 | **FHIR:** R5 | **Database:** PostgreSQL 18.3
> **Status:** 🔴 Bootstrap

R4↔R5 translation layer — bridges partner systems still on R4

## What This Service Does

This is a Go-based FHIR R5 microservice that is part of the ZarishSphere digital health platform.
It stores all clinical data as FHIR R5 resources in PostgreSQL 18.3 (JSONB format),
communicates via NATS 2.12.5 JetStream events, and is secured with SMART on FHIR 2.1 via Keycloak 26.5.6.

## Repository Structure

```
zs-core-fhir-r4-bridge/
├── cmd/
│   └── server/          # main.go — service entry point
├── internal/
│   ├── api/             # HTTP handlers, router (chi v5)
│   ├── fhir/            # FHIR resource operations
│   ├── storage/         # PostgreSQL 18.3 layer (pgx v5.7.x)
│   ├── events/          # NATS 2.12.5 JetStream events
│   └── auth/            # SMART on FHIR scope validation
├── migrations/          # SQL migration files
├── config/              # Configuration (viper)
├── deploy/
│   ├── Dockerfile
│   └── helm/            # Helm 3.17+ chart
├── .github/
│   └── workflows/       # CI/CD pipelines
├── Makefile
├── go.mod
└── README.md
```

## Technology Stack

| Component | Version |
|-----------|---------|
| Go | **1.26.1** |
| PostgreSQL | **18.3** |
| NATS JetStream | **2.12.5** |
| Valkey (cache) | **9.0.3** |
| Keycloak (auth) | **26.5.6** |
| OpenTelemetry | **1.35+** |

## Getting Started (Local Development)

**Prerequisites:** Ensure you have run `bootstrap.sh` from the ZarishSphere platform root.
This starts your full local dev stack automatically.

```bash
# 1. Clone this repository
git clone https://github.com/zarishsphere/zs-core-fhir-r4-bridge.git
cd zs-core-fhir-r4-bridge

# 2. Run tests
make test

# 3. Start the service locally
make dev
```

The service will be available at `http://localhost:PORT/fhir/R5/`

## API Reference

See [docs.zarishsphere.com/api/zs-core-fhir-r4-bridge](https://docs.zarishsphere.com/api/zs-core-fhir-r4-bridge) for the full OpenAPI 3.1 specification.

## Contributing

All contributions must go through our RFC process. See [CONTRIBUTING.md](https://github.com/zarishsphere/.github/blob/main/CONTRIBUTING.md).

---

*Part of the [ZarishSphere Platform](https://zarishsphere.com) — Free · Open Source · FHIR R5 · Offline-First*
