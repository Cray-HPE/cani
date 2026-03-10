## cani alpha add cable

Add cable(s) to the inventory.

### Synopsis

Add one or more cables to the inventory by slug or part number.

```
cani alpha add cable <slug-or-part-number> [flags]
```

### Options

```
      --a-device string   Termination A device UUID or name
      --a-port string     Termination A port name
      --b-device string   Termination B device UUID or name
      --b-port string     Termination B port name
      --color string      Cable color
  -h, --help              help for cable
      --label string      Cable label
```

### Options inherited from parent commands

```
  -y, --accept                 Automatically accept recommended values.
  -a, --auto                   Automatically recommend values for parent hardware
      --config string          config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string       datastore type (json, postgres) (default "json")
      --debug                  enable debug mode
  -L, --list-supported-types   List supported hardware types.
  -p, --parent string          Parent item UUID. (default "00000000-0000-0000-0000-000000000000")
  -q, --qty int                Quantity of items to add. (default 1)
      --strict                 require a resolved device type (slug) for all devices (default true)
      --types-dirs strings     local directories with additional hardware types
      --types-repo-pull        pull latest changes from types repos on startup
      --types-repos strings    git repo URLs with additional hardware types
```

### SEE ALSO

* [cani alpha add](cani_alpha_add.md)	 - Add items to the inventory

