# Extending The Device Types Library

`cani` loads device types from three sources, in priority order:

1. **Built-in** — embedded types shipped with the binary.
2. **Local directories** — YAML files on disk (`types_dirs`).
3. **Remote git repos** — cloned repositories (`types_repos`).

Types registered first win. If a slug already exists from a higher-priority source, the duplicate is skipped.

## Configuration

All type sources are configured in `~/.cani/cani.yml`:

```yaml
# Local directories containing device type YAML files.
# Each directory should have subdirectories like device-types/, rack-types/, etc.
types_dirs: []

# Git repositories to clone for additional device types.
# The default is the netbox-community device-type library.
types_repos:
  - https://github.com/netbox-community/devicetype-library.git

# Whether to git pull on each run (false = clone once, never update).
types_repo_pull: false
```

## Adding A Local Directory

Create a directory with the standard layout and add it to `types_dirs`:

```
my-types/
├── device-types/
│   └── my-custom-blade.yaml
├── rack-types/
│   └── my-custom-rack.yaml
├── module-types/
│   └── my-custom-gpu.yaml
└── cable-types/
    └── my-custom-cable.yaml
```

```yaml
types_dirs:
  - /home/user/my-types
```

Types from local directories take priority over remote repos, so you can override community types with local definitions.

## Adding A Git Repository

Add a repository URL to `types_repos`:

```yaml
types_repos:
  - https://github.com/netbox-community/devicetype-library.git
  - https://github.com/my-org/custom-device-types.git
```

Repos are cloned into `~/.cani/types/<sanitized-repo-name>/` on first use. Set `types_repo_pull: true` to pull updates on each run.

## Listing Loaded Types

After configuring sources, use the `-L` flag to verify which types are available:

```shell
# List all loaded device types
cani alpha add device -L

# List all loaded rack types
cani alpha add rack -L

# List all loaded module types
cani alpha add module -L

# List all loaded cable types
cani alpha add cable -L
```

The output includes slugs from all sources (built-in, local, and remote):

```
rack (2):
NAME                                 SLUG                                           PART NUMBER       SOURCE
-----------------------------------  ---------------------------------------------  ----------------  ----------
HPE 42U 800mmx1200mm G2 Enterprise…  hpe-42u-800mmx1200mm-g2-enterprise-shock-rack  P9K46A            builtin
HPE 48U 800mmx1200mm G2 Enterprise…  hpe-48u-800mmx1200mm-g2-enterprise-shock-rack  P9K58A            builtin

Cabinet (7):
NAME                                 SLUG                                           PART NUMBER       SOURCE
-----------------------------------  ---------------------------------------------  ----------------  ----------
EX2000                               hpe-eia-cabinet                                                  builtin
EX2000                               hpe-ex2000                                                       builtin
EX2500 1 Liquid Cooled Chassis       hpe-ex2500-1-liquid-cooled-chassis                               builtin
EX2500 2 Liquid Cooled Chassis       hpe-ex2500-2-liquid-cooled-chassis                               builtin
EX2500 3 Liquid Cooled Chassis       hpe-ex2500-3-liquid-cooled-chassis                               builtin
EX3000                               hpe-ex3000                                                       builtin
EX4000                               hpe-ex4000                                                       builtin
```

Use a slug from this list with any `add` command:

```shell
cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --auto --accept
```

## Priority Order Example

If the slug `my-blade` exists in both a local directory and a git repo:

1. Built-in types load first — no match.
2. Local `types_dirs` loads `my-blade` — **registered**.
3. Git `types_repos` finds `my-blade` — **skipped** (already registered).

This lets you fork and customize community types without modifying the upstream repository.

## Subdirectory Layout

Each source directory is scanned for these subdirectories:

| Subdirectory | Type Loaded |
|--------------|-------------|
| `device-types/` | `CaniDeviceType` |
| `module-types/` | `CaniModuleType` |
| `cable-types/` | `CaniCableType` |
| `rack-types/` | `CaniRackType` |
| `inventory-types/` | `CaniFruType` |

Any `.yaml` or `.yml` file in these subdirectories is parsed and registered by slug.
