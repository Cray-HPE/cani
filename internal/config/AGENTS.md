# Config Package

## Purpose

Configuration management for CANI with YAML comment and key ordering preservation during round-trip read/write operations.

## Layout

| File | Purpose |
|------|---------|
| `config.go` | Core load/save logic with YAML node tree preservation |
| `comments.go` | Comment metadata registry using struct tag reflection |
| `migrate_test.go` | Tests for migrating deprecated config formats |

## Key Types

- `Cfg` — Main configuration struct; holds provider settings, paths, and raw YAML node tree
- `CommentRegistry` — Stores comment metadata for provider fields
- `FieldComment` — Head, line, and foot comments for a field

## Key Functions

- `LoadOrCreate(path)` — Reads or creates config; preserves YAML structure
- `Save()` — Writes config back, preserving comments and ordering
- `GetNestedValue/String/Int` — Safe nested value access with defaults
- `RegisterProviderComments` — Extracts comment metadata from struct tags

## Patterns

- Singleton `Config` variable for global access
- YAML node tree preservation (not just marshal/unmarshal)
- Defensive merging: only adds missing keys, never overwrites user values
- Provider interface polymorphism (`DefaultOptionsProvider`, etc.)
