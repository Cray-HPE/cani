## Critical (non-negotiable)

- provider-specific rack/device staging logic belongs here, not in `cmd/`
- CSM uses the term "cabinet" which maps to cani's "rack" (`CaniRackType`)
- providers implement optional interfaces from `internal/provider/` to hook into the generic command layer (e.g. `RackPostAddHook` for post-add rack logic)

## Do

### Keep Extract, Transform, and Load operations distinctly separate

- when importing, save raw data to the provider struct
- when transforming, parse raw data from provider struct or from cani inventory
- use getters and setters to get info to and from the provider struct to keep logic isolated to each sub package

### Emit shared types the merge can validate

- set `ModuleBayName` only to a bay the parent device type declares; the merge canonicalises the
  spelling, clears an undeclared bay and reports it, so a provider-specific placement descriptor
  belongs in `ProviderMetadata`, not in the bay name (see the Transform output contract in
  `internal/provider/AGENTS.md`)
- when the source system declares a device's module bays, set `CaniDeviceType.ModuleBays` from
  them (the Nautobot provider's `ApplyModuleBays`), so a bay name the library spells differently
  is declared for that device instead of being cleared by the merge
- leave out a record you cannot translate and list it in `TransformResult.Warnings` rather than
  inventing a value for it

## Don't

- do not repeat flags that exist in parent or root commands
- parse common flags from parent or root commands (debug, merge, step, etc.)
