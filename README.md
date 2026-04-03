# zs-core-fhir-r4-bridge

> **ZarishSphere Platform** · [github.com/orgs/zarishsphere](https://github.com/orgs/zarishsphere)

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](../../LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go)](https://golang.org)
[![FHIR R5](https://img.shields.io/badge/FHIR-R5%205.0.0-orange)](https://hl7.org/fhir/R5/)
[![CI](https://github.com/zarishsphere/zs-core-fhir-r4-bridge/actions/workflows/ci.yml/badge.svg)](https://github.com/zarishsphere/zs-core-fhir-r4-bridge/actions)

Bidirectional FHIR R4 (4.0.1) ↔ R5 (5.0.0) translation bridge. Enables ZarishSphere's R5-native platform to interoperate with partner systems (DHIS2, OpenMRS, legacy HIS) still on FHIR R4.

---

## Quick start

```bash
# Run locally (requires Go 1.26.1)
make dev

# Run tests
make test

# Build binary
make build

# Build multi-arch Docker image (amd64 + arm64 / Raspberry Pi 5)
make docker-build
```

## API

| Path | Method | Description |
|------|--------|-------------|
| `/healthz` | GET | Liveness probe |
| `/readyz` | GET | Readiness probe |
| `/metrics` | GET | Prometheus metrics |

Listening on port **8085** by default. Override with `SERVER_ADDR=:PORT`.

---

**Part of ZarishSphere** · Apache 2.0 · Free forever · [platform@zarishsphere.com](mailto:platform@zarishsphere.com)
