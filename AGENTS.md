# AGENTS.md

## Critical (non-negotiable)

- the `cmd/` layer MUST NOT contain provider-specific logic; it orchestrates generic `devicetypes` operations only
- provider-specific logic (CSM, HPCM, Redfish, etc.) belongs in `pkg/provider/<name>/`
- providers hook into the command layer via interfaces defined in `internal/provider/`

## Do

- use the Standard Go Project Layout
- default to small components
- default to small diffs
- keep cognitive complexity < 10 per function
- file size < 300 lines
- write modular, maintainble code
- write functions that do one thing and one thing well

## Don't

- do not use third-party libraries
- do not hard code variables
- do not import provider packages from `cmd/`

## File Organization

## Commands

### Building

```bash
make bin              # build the binary
```
