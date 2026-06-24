---
name: code-review
description: "Senior Architect review of cani: layered architecture, data/responsibility separation, the provider plugin system, and the portable Nautobot-parity data model that generalizes CRUD"
argument-hint: "[--scope=full|branch|file] [--pillar=layers|data|providers|model|cross]"
allowed-tools: Read, Bash, Glob, Grep, Task
---

# Code Review Skill

Perform a comprehensive **Senior Architect** level review of the cani codebase. This is
not a line-by-line lint pass — it judges whether the *system design* holds up. The review
is anchored on the four properties that define this project's long-term health, plus a
lighter cross-cutting pass.

Ground every finding in the boundary rules the repo already declares for itself:
`AGENTS.md` (root), `cmd/AGENTS.md`, `internal/provider/AGENTS.md`, and
`pkg/provider/AGENTS.md`. Treat a violation of those rules as a concrete defect, not a
matter of taste. Cite files and line numbers; show the offending import or call.

## The map (orient here first)

```
main.go                  entrypoint: blank-imports providers (self-registration), Init() + Execute()
cmd/                     orchestration / CLI verbs (add remove update show import export classify serve init)
internal/cli/            stdlib command framework (Command/FlagSet) — replaced cobra/viper/pflag
internal/provider/       PLUGIN CONTRACTS: interface.go (1 required + 11 optional ifaces) + registry.go
internal/config/         configuration (YAML node tree, env precedence, Cfg.Providers sub-maps)
internal/core/           app constants
pkg/provider/<name>/     provider implementations: csm hpcm nautobot ochami redfish example
pkg/devicetypes/         PORTABLE MODEL (Nautobot/Netbox-shaped) + hardware-type library: Inventory, Cani*Type, TransformResult, CaniType
pkg/devicetypes/NAUTOBOT_MAPPING.md   authoritative Cani*Type <-> Nautobot field mapping & coverage tracker
pkg/datastores/          persistence abstraction (DeviceStore; JSON store today)
pkg/visual/              presentation (tables, trees, rack views) — read-only over the model
pkg/xname/ pkg/nautobot/ pkg/canu/   provider-adjacent helpers / external clients
```

Intended dependency direction is **inward**: `cmd/` → `internal/cli` + `internal/provider`
(contracts) + `pkg/devicetypes`/`pkg/datastores`; providers depend on the contracts and the
shared model, never the reverse; the shared model depends on nothing app-specific.

## Pillars

### 1. Layered architecture & dependency direction  (`--pillar=layers`)

Verify the layers above are real, acyclic, and pointing inward — no shortcuts, no leakage.

- **Inspect:** `go list -deps ./cmd/... | grep provider`; `grep -rn "pkg/provider/" cmd/ main.go`;
  `grep -rn "Cray-HPE/cani/cmd" pkg/ internal/`; read each `AGENTS.md`.
- **Look for:** `cmd/` reaching providers *only* through `internal/provider` interfaces and the
  blank-import self-registration in `main.go`; shared packages (`pkg/devicetypes`,
  `pkg/datastores`, `internal/cli`, `internal/config`) that import neither `cmd/` nor any
  provider; a clean one-way graph.
- **Red flags:** any `import ".../pkg/provider/<name>"` inside `cmd/`; `switch providerName`
  or provider-name string literals in `cmd/`; the data model or datastore importing a
  provider; import cycles; files >300 lines or functions with cognitive complexity >10
  (the repo's stated budgets); business logic living inside `internal/cli`.

### 2. Separation of data & responsibility  (`--pillar=data`)

Each package should own exactly one responsibility, and the inventory **data** must stay
decoupled from transport, persistence, configuration, and rendering.

- **Inspect:** `grep -rnE "net/http|internal/cli|spf13" pkg/devicetypes/`;
  `grep -rn "provider" pkg/datastores/`; review `pkg/devicetypes` types, `pkg/datastores`
  `DeviceStore`, `internal/config` save path, and how `--metadata`/`--tag` flow from `cmd/`
  into providers.
- **Look for:** `pkg/devicetypes` as pure domain types + library loading (no HTTP, no CLI,
  no provider specifics); `pkg/datastores` provider-agnostic behind `DeviceStore` so the
  backend can be swapped; config persistence that is lossless (comments/key order preserved
  via the YAML node tree) and writes token-bearing files `0600`; generic CLI concerns parsed
  in `cmd/` then *handed* to providers (e.g. `MetadataApplier`) so no provider name is
  hard-coded; the ETL stages **Import (Extract) → Transform → Load (Export)** kept distinct,
  with the provider holding extracted state between stages.
- **Red flags:** provider-specific fields baked into the shared model (e.g. xname/auth
  concepts in `pkg/devicetypes`); a single function or file that mixes flag parsing +
  business logic + rendering + persistence; rendering code in `pkg/visual` mutating the
  inventory; config option structs conflated with CLI flag wiring; secrets logged or
  world-readable.

### 3. Provider plugin architecture  (`--pillar=providers`)

The extensibility core. Judge the contract, registration, command injection, capability
dispatch, and how a new provider is onboarded.

- **Inspect:** `internal/provider/interface.go` (required `Provider` = `Transform` /
  `NewProviderCmd` / `Slug`, plus the 11 optional capability interfaces: `Exporter`,
  `Importer`, `HasOptions`, `HasImportOptions`, `HasExportOptions`, `DeviceStager`,
  `RackStager`, `RackPostAddHook`, `MetadataApplier`, `DeviceUpdateFlagProvider`,
  `StagedDeviceDescriber`); `internal/provider/registry.go` (`Register`/`GetProvider`/
  `GetProviders`); the `NewProviderCmd(base)` call sites and RunE-wrapping in
  `cmd/import` and `cmd/export`; the optional-interface type-asserts in `cmd/`
  (`grep -rn "p.(provider\." cmd/` and `GetProviders()` loops); the `cani init` scaffolder.
- **Look for:** every integration point reached through an interface or a type-assert on an
  optional interface — never a provider name; the down-and-back-up command mechanism (cmd/
  passes `*cli.Command` to `NewProviderCmd`, provider switches on `base.Name()` and returns
  a customized command that cmd/ wires back into the tree); self-registration via `init()` +
  blank imports; a coherent required-vs-optional split so providers implement only what they
  support; each provider keeping its specifics (xname decode, auth, API clients) inside
  `pkg/provider/<name>` and its `pkg/<name>` client; uniform dry-run / idempotency / error
  semantics across providers; graceful degradation when a provider returns a nil command or
  a `Configure` error.
- **Red flags:** "god" interfaces forcing unused methods, or interface explosion that makes
  a new provider hard to write; provider behavior selected by string switch in shared code;
  duplicate-registration or concurrency hazards in the registry; ETL semantics that differ
  silently per provider; provider packages importing each other; missing test seam (the
  `fakeProvider` pattern in `registry_test.go`) for contract conformance.

### 4. Portable data model, Nautobot parity & provider-agnostic CRUD  (`--pillar=model`)

The linchpin of the whole design. `pkg/devicetypes` defines a **portable, Nautobot/Netbox-
shaped** schema that is the single contract everything else generalizes over. Every item is
UUID-keyed in `Inventory` (Locations, Racks, Devices, Modules, Cables, Frus, Interfaces,
plus IPAM Prefixes/IPAddresses/VLANs and Metadata) and implements `CaniType`
(`Validate`/`GetID`/`GetSlug`/`GetStatus`). Because **CRUD operates only on this model**,
`add`/`remove`/`update`/`show` are byte-for-byte identical for every provider — providers
translate **only** at the edges (Import/Transform/Export). `pkg/devicetypes/NAUTOBOT_MAPPING.md`
is the authoritative field-by-field mapping and the parity/coverage tracker; treat it as part
of the schema's source of truth.

- **Inspect:** read `pkg/devicetypes/inventory.go` (`Inventory`, `TransformResult`,
  `NewInventory`, `EnsureUniqueDeviceNames`), `cani_type.go` (`CaniType`), `registry.go`
  (`ClassifyForNautobot`), `inventory_relationships.go` (`VerifyParentChildRelationships`),
  and `NAUTOBOT_MAPPING.md`. Then
  `grep -rniE "redfish|csm|hpcm|nautobot|ochami|provider" cmd/add cmd/remove cmd/update cmd/show pkg/devicetypes`
  and `grep -rn "ProviderMetadata" pkg/provider/`.
- **Look for:**
  - **Provider-agnostic CRUD** — `cmd/add|remove|update|show` operate purely on the
    `Inventory`/`Cani*Type` through `pkg/datastores`; no provider import, no `switch
    provider`, no provider-named branch. Adding hardware resolves a **slug** against the
    embedded device-type library so template fields (interfaces, module bays, power ports)
    auto-populate — never hand-coded specs.
  - **Nautobot parity** — any new or changed `Cani*Type` field has a corresponding entry in
    `NAUTOBOT_MAPPING.md` (Mapped / Partial / Not-Mapped / Cani-Internal), respects the
    **template-vs-instance** split (library/template fields → Nautobot *type* objects;
    instance fields → Nautobot instances), and follows the **NetBox devicetype-library tag
    casing**: kebab-case component keys (`console-ports`, `power-ports`, `module-bays`,
    `device-bays`) with snake_case NetBox scalars and **camelCase JSON**. New `Type`
    constants are routed in `ClassifyForNautobot()`.
  - **Relationship model mirrors Nautobot** — relationships are single-direction child FKs
    (`Parent`, `Location`, `ParentDevice`, `Device`, cable terminations); reverse indices
    (`Children`, `Racks`, `Devices`, `OccupiedSlots`, `Interfaces`) are **rebuilt** by
    `VerifyParentChildRelationships()` and marked non-serialized, never hand-set and
    persisted. Parent-chain cycle detection is preserved.
  - **Provider escape hatch** — provider-only attributes live in `ProviderMetadata`
    (exported as Nautobot CustomFields), not as new columns on the shared types. `Transform`
    emits a `TransformResult` of shared types and calls `EnsureUniqueDeviceNames`.
  - **Lossless round-trip** — the datastore serializes/deserializes the portable schema
    without loss; transient caches (e.g. `pkIndex`) are `yaml:"-" json:"-"` and rebuilt after
    load, not persisted.
- **Red flags:**
  - CRUD that branches on, or imports, a provider, or re-implements add/update logic per
    provider instead of operating on the portable model.
  - A provider inventing its own inventory shape instead of producing
    `TransformResult`/`Cani*Type`, or writing provider-specific fields onto a shared type
    instead of `ProviderMetadata`.
  - A new shared field with **no** `NAUTOBOT_MAPPING.md` entry, a YAML tag that breaks
    NetBox devicetype-library parity (e.g. snake_casing a kebab-case component key such as
    `console-ports`/`module-bays`), or a missed template/instance classification.
  - Reverse-relationship pointers persisted to disk or maintained by hand; a new `Cani*Type`
    that does not implement `CaniType`; a new hardware category not wired into
    `ClassifyForNautobot()`.
  - Bypassing the slug/device-type library (hard-coding device specs inside a provider) so
    the same hardware diverges between providers.

### Cross-cutting pass  (`--pillar=cross`)  — supporting, lighter weight

Confirm the four pillars are backed by: the multi-tier test strategy (Go unit tests +
shellspec `spec/functional` and `spec/integration`, run via `make utest|ftest|itest`); the
**standard-library-only** constraint (no third-party deps reintroduced — `grep spf13 go.mod`
must be empty); conventional-commit + signed-commit hooks and CI; MIT license headers in
scanned dirs; and cross-architecture builds. Flag gaps only insofar as they undermine the
layering, data separation, plugin contract, or portable-model parity.

## Output

For each of the four core pillars give a score (1–5) with evidence, and a short
cross-cutting note. Then:

- **Boundary-violation ledger** — every concrete breach of an `AGENTS.md` rule or an
  established pattern (provider leakage into `cmd/`, model/persistence coupling, broken
  dependency direction, hard-coded provider names, **CRUD that branches on a provider**,
  **provider-specific fields on shared `Cani*Type`s instead of `ProviderMetadata`**, and
  **`Cani*Type` changes that drift from `NAUTOBOT_MAPPING.md` or the NetBox-parity tag
  casing (kebab-case component keys)**), with file:line and the fix.
- **Overall verdict** — Junior / Mid-Level / Senior / Staff+ engineering standard, justified.
- **Top-5 improvements** ranked by impact on extensibility and maintainability.

Be honest and specific; prefer "this import in `cmd/x.go:42` breaks the cmd→provider
boundary" over generalities.

## Process

1. Spawn parallel explore agents — one per core pillar (layers, data, providers, model) plus
   one for the cross-cutting pass. Give each the map above and the relevant `AGENTS.md` paths.
2. Each agent returns evidence (file:line, import graphs, offending snippets), not opinions.
3. Synthesize into the structured report; assign per-pillar scores and the boundary-violation
   ledger.
4. Produce the overall verdict and the ranked top-5 improvements.


## Sub-agent worktree safety

When dispatching ANY sub-agent (`Task`) that reads or writes repository files:

1. **Pin the absolute worktree path** in the prompt (the exact directory the
   agent must operate in) and instruct it to run every command from there.
2. **Forbid sibling worktrees by name** — explicitly tell the agent to ignore
   any other `origin*` checkout, and name the known stale one(s) if relevant.
3. **Require a pre-flight proof.** Have the agent run, before any real work:

   ```bash
   cd <ABSOLUTE_WORKTREE_PATH> && git rev-parse --short HEAD && git branch --show-current
   ```

   Confirm it matches the expected SHA/branch. Add a content tripwire when
   useful (e.g. `grep -c <expected-symbol> <file>` must be non-zero) so the
   agent proves it is on the tree that actually contains the work.
4. **STOP-and-report** if the pre-flight does not match — never let the agent
   "adapt" to an unexpected tree.
5. On return, **verify the agent's commits landed in the intended worktree**
   (`git -C <ABSOLUTE_WORKTREE_PATH> log --oneline -n 20`), not a sibling.
