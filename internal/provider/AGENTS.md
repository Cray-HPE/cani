# Provider Package

## Purpose

The provider package is the **foundational interface** for CANI. All inventory providers must implement this interface to integrate with the system. Providers follow an **ETL (Extract-Transform-Load)** pattern for syncing external inventory sources with CANI's internal datastore.

## Layout

| File | Purpose |
|------|---------|
| `interface.go` | Core `Provider` interface and optional extension interfaces |
| `struct_to_map.go` | Reflection utility for converting structs to maps |

## Core Interface

Every provider must implement `Provider`:

| Method | Purpose |
|--------|---------|
| `Transform` | Convert external data to CANI format (Transform step) |
| `NewProviderCmd` | Return provider-specific CLI commands |
| `Slug` | Return provider identifier |

## Optional Interfaces

Providers can implement these for additional capabilities:

| Interface | Purpose |
|-----------|---------|
| `Importer` | Extract data from external systems (Extract step`) |
| `Exporter` | Sync local inventory to external system (Load step) |
| `HasOptions` | Expose default configuration options |
| `HasImportOptions` | Import-specific config/flags |
| `HasExportOptions` | Export-specific config/flags |

## Provider Registry

- `Register(name, provider)` — Called in provider's `init()` to register
- `GetProvider(name)` — Retrieve a registered provider
- `GetProviders()` — Get all registered providers

## Implementing a New Provider

1. Create a new package under `pkg/provider/<name>/` via `cani init <new-provider-name>`
2. Implement the `Provider` interface
3. Call `provider.Register("<name>", &YourProvider{})` in `init()`
4. Implement optional interfaces as needed (`Importer`, `HasOptions`, etc.)
5. Keep methods focused—each should do one thing well
