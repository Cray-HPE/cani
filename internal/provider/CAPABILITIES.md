# Provider capability matrix

`cani`'s plugin system is built from one **required** `Provider` contract
(`Transform`, `NewProviderCmd`, `Slug`) plus a set of **optional** capability
interfaces declared in [`interface.go`](interface.go). A provider implements only
the interfaces it supports; the command layer type-asserts each one and degrades
gracefully when it is absent.

The table below records which optional interface each registered provider
implements. It is **verified and regenerated** by the conformance suite in
[`conformance/conformance_test.go`](conformance/conformance_test.go) — run
`go test ./internal/provider/conformance/... -v` to print the live matrix. Keep
this table in sync when a provider gains or drops a capability.

| Optional interface | csm | example | hpcm | nautobot | ochami | redfish |
|--------------------|:---:|:-------:|:----:|:--------:|:------:|:-------:|
| `Importer` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `Exporter` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `HasOptions` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `Configurer` |  |  |  |  |  |  |
| `HasImportOptions` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `HasExportOptions` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `DeviceStager` | ✓ |  |  |  |  |  |
| `RackStager` | ✓ |  |  |  |  |  |
| `RackPostAddHook` | ✓ |  |  |  |  |  |
| `MetadataApplier` |  |  |  | ✓ |  |  |
| `DeviceUpdateFlagProvider` | ✓ |  |  |  |  |  |
| `StagedDeviceDescriber` | ✓ |  |  |  |  |  |

## Notes

- Every provider implements the **ETL core** (`Importer`, `Exporter`,
  `HasOptions`, `HasImportOptions`, `HasExportOptions`), so import/export and
  their config/flags work uniformly.
- The richer **add/update staging** capabilities (`DeviceStager`, `RackStager`,
  `RackPostAddHook`, `DeviceUpdateFlagProvider`, `StagedDeviceDescriber`) are
  CSM-specific today; they keep xname/cabinet logic inside the CSM provider
  rather than in `cmd/`.
- `MetadataApplier` is implemented only by Nautobot, which maps generic
  `--metadata` pairs into its own sub-map.
- `Configurer` is currently implemented by **no** provider. The interface exists
  for startup config validation but is unused — a candidate to either adopt or
  remove.
