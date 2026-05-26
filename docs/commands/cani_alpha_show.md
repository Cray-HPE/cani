## cani alpha show

Show items from the inventory

### Synopsis

Show items from the inventory.

```
cani alpha show [flags]
```

### Options

```
      --cable-type string   Filter cables by type (e.g., 'dac', 'cat6', used with --rack-view)
      --columns int         Number of rack columns before wrapping (0=auto, used with --rack-view)
  -f, --file string         Load inventory from YAML file (used with --visual)
  -o, --format string       Output format (json) (default "json")
  -h, --help                help for show
      --no-color            Disable colorized output (used with --visual)
      --rack string         Filter to specific rack by name (used with --visual)
      --rack-view           Display compact ASCII rack view with device symbols
      --show-cables         Show cable connections in visual output
      --show-routing        Display cable routing with branching visualization (1 rack per line)
  -s, --sort string         Sort by this field (name, type, id, status, vendor, model) (default "name")
  -V, --verbose count       Verbose output: -V shows legend, -VV shows all cables (used with --rack-view)
  -v, --visual              Display ASCII visualization of rack layout
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
* [cani alpha show cable](cani_alpha_show_cable.md)	 - List cables in the inventory.
* [cani alpha show cables](cani_alpha_show_cables.md)	 - List cables in the inventory.
* [cani alpha show device](cani_alpha_show_device.md)	 - List devices in the inventory.
* [cani alpha show interfaces](cani_alpha_show_interfaces.md)	 - List interfaces in the inventory.
* [cani alpha show location](cani_alpha_show_location.md)	 - List locations in the inventory.
* [cani alpha show module](cani_alpha_show_module.md)	 - List modules in the inventory.
* [cani alpha show rack](cani_alpha_show_rack.md)	 - List racks in the inventory.

