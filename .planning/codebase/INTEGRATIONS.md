# External Integrations

**Analysis Date:** 2026-04-30

## APIs & External Services

**CSM (Cray System Management):**
- SLS (System Layout Service) - Hardware topology/location management
  - Client: Custom HTTP client (`pkg/provider/csm/client/client.go`)
  - Base URL: `https://<host>/apis/sls/v1`
  - Auth: Keycloak OAuth2 password-credentials grant
  - Config: `providers.csm.provider_host`, token credentials

- HSM (Hardware State Manager) - Hardware state and inventory
  - Client: Custom HTTP client (`pkg/provider/csm/client/client.go`)
  - Base URL: `https://<host>/apis/smd/hsm/v2`
  - Auth: Same Keycloak token as SLS

**Nautobot (Network Source of Truth):**
- Full DCIM API integration for device/location/module management
  - SDK: Generated OpenAPI client (`pkg/nautobot/nautobot_api.go`)
  - Generator: `oapi-codegen` v2.5.1
  - Auth: Token-based (`Token <api-token>` header)
  - Client wrapper: `pkg/provider/nautobot/export/client.go`
  - Operations: Device CRUD, location types, module types, roles, statuses

**OpenCHAMI (Open Composable HPC Architecture Management Interface):**
- JSON inventory import/export
  - Client: `pkg/provider/ochami/` (file-based, imports JSON inventory)
  - Auth: Not detected (file-based workflow)

**Redfish (DMTF standard):**
- Hardware discovery from Redfish ServiceRoot
  - Client: `pkg/provider/redfish/` (reads JSON files or stdin)
  - Auth: Not detected (file-based import)

**HPCM (HPC Cluster Manager):**
- Cluster inventory import from cmdb/cmconfig
  - Client: HTTP GET for cmdb data (`pkg/provider/hpcm/import/cmdb_parse.go`)
  - Config formats: INI (`cmconfig`), HTTP endpoints (`cmdb`)

**CANU (CSM Automatic Network Utility):**
- Paddle file (network topology) import
  - Client: Local file parsing (`pkg/canu/canu.go`)
  - Format: JSON "paddle" files with topology data

## Data Storage

**Databases:**
- JSON file datastore (current implementation)
  - Interface: `pkg/datastores/datastore.go` — `DeviceStore` interface
  - Implementation: `NewJSONStore()` — local JSON file
  - Config: `datastore` field in config YAML
- PostgreSQL (planned, not yet implemented)
  - Defined as `StoreTypePostgres` constant, code commented out

**File Storage:**
- Local filesystem only
  - Config file: YAML in user config directory
  - Inventory: JSON file (path from config `datastore` field)
  - Device types: YAML files in `pkg/devicetypes/` and configurable `types_dirs`

**Caching:**
- None (all state persisted to JSON file)

## Authentication & Identity

**CSM Provider Auth:**
- Keycloak OAuth2 password-credentials grant
  - Token URL: `https://<host>/keycloak/realms/shasta/protocol/openid-connect/token`
  - Client ID: `shasta`
  - Scope: `openid`
  - Implementation: `pkg/provider/csm/client/auth.go`
  - Options: Pre-existing bearer token OR username/password credentials
  - TLS: Custom CA cert support, optional `InsecureSkipVerify`

**Nautobot Provider Auth:**
- API Token authentication
  - Header: `Authorization: Token <token>`
  - Implementation: `pkg/provider/nautobot/export/client.go`

**CSM Simulation Mode:**
- Skips auth, uses `localhost:8443`, forces `InsecureSkipVerify`
- Config: `UseSimulation` option in `pkg/provider/csm/client/options.go`

## Monitoring & Observability

**Error Tracking:**
- None (CLI tool — errors returned to user)

**Logs:**
- Go `log` standard library
- Custom prefix logging: `[datastores]`, etc.
- `--debug` flag enables verbose output
- Nautobot provider has color logging (`pkg/provider/nautobot/logcolor/`)

## CI/CD & Deployment

**Hosting:**
- GitHub (Cray-HPE/cani)
- RPM package distribution (`cani.spec`)

**CI Pipeline:**
- GitHub Actions (`.github/workflows/`)
  - `unit_tests.yml` - Go unit tests on push/PR
  - `shellspec.yml` - ShellSpec BDD tests
  - `shellcheck.yml` - Shell script linting
  - `license_check.yml` - License header verification
  - `promote-prerelease.yml` / `promote-release.yml` - Release promotion

**Documentation:**
- MkDocs with Material theme (`mkdocs.yml`, `docs/`)
- Mike for doc versioning

## Environment Configuration

**Required env vars (CSM provider):**
- Provider host (via config or flag)
- Keycloak username/password OR pre-existing API gateway token
- Optional: CA cert path, K8s pod/service CIDRs

**Required env vars (Nautobot provider):**
- Nautobot URL (via config)
- Nautobot API token (via config)

**Secrets location:**
- Not stored in files — passed via CLI flags or config file
- CSM: Keycloak credentials or bearer token
- Nautobot: API token in config
- K8s secrets supported (`SecretName` option)

## Webhooks & Callbacks

**Incoming:**
- None (CLI tool; `serve` command exists but is not yet implemented — `cmd/serve/serve.go`)

**Outgoing:**
- None

## Device Type Library

**External Repository:**
- Default: `https://github.com/netbox-community/devicetype-library.git`
  - Configured in `internal/config/config.go` as `DefaultTypesRepo`
  - Cloned locally when `types_repo_clone: true`
  - Auto-pulled when `types_repo_pull: true`

**Embedded Types:**
- Hardware types embedded in `pkg/devicetypes/` (YAML definitions)
- Vendors: Cisco, CrayGigabyte, F5-Networks, Fortinet, HPE, Motivair, NetApp, NVIDIA, Raritan

---

*Integration audit: 2026-04-30*
