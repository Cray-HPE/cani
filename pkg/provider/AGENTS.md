## Critical (non-negotiable)

- provider-specific rack/device staging logic belongs here, not in `cmd/`
- CSM uses the term "cabinet" which maps to cani's "rack" (`CaniRackType`)
- providers implement optional interfaces from `internal/provider/` to hook into the generic command layer (e.g. `RackPostAddHook` for post-add rack logic)

## Do

### Keep Extract, Transform, and Load operations distinctly separate

- when importing, save raw data to the provider struct
- when transforming, parse raw data from provider struct or from cani inventory
- use getters and setters to get info to and from the provider struct to keep logic isolated to each sub package

## Don't

- do not repeat flags that exist in parent or root commands
- parse common flags from parent or root commands (debug, merge, step, etc.)
