## cani alpha add

Add items to the inventory

### Synopsis

Add items to the inventory.

When called with a slug or part number, searches all hardware registries
(rack, device, module, cable) and automatically determines the type.

Use subcommands (rack, device, module, cable, location) to constrain
to a specific type; subcommands reject slugs that do not match their type.

```
cani alpha add [slug-or-part-number] [flags]
```

### Options

```
  -y, --accept                 Automatically accept recommended values.
  -a, --auto                   Automatically recommend values for parent hardware
  -h, --help                   help for add
  -L, --list-supported-types   List supported hardware types.
  -p, --parent string          Parent item UUID. (default "00000000-0000-0000-0000-000000000000")
  -q, --qty int                Quantity of items to add. (default 1)
```

### Options inherited from parent commands

```
      --config string         config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string      datastore type (json, postgres) (default "json")
      --debug                 enable debug mode
      --strict                require a resolved device type (slug) for all devices (default true)
      --types-dirs strings    local directories with additional hardware types
      --types-repo-pull       pull latest changes from types repos on startup
      --types-repos strings   git repo URLs with additional hardware types
```

### SEE ALSO

* [cani alpha](cani_alpha.md)	 - Run commands that are considered unstable.
* [cani alpha add cable](cani_alpha_add_cable.md)	 - Add cable(s) to the inventory.
* [cani alpha add device](cani_alpha_add_device.md)	 - Add device(s) to the inventory.
* [cani alpha add location](cani_alpha_add_location.md)	 - Add a location to the inventory.
* [cani alpha add module](cani_alpha_add_module.md)	 - Add module(s) to the inventory.
* [cani alpha add rack](cani_alpha_add_rack.md)	 - Add rack(s) to the inventory.

