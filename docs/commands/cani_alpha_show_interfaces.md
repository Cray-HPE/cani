## cani alpha show interfaces

List interfaces in the inventory.

### Synopsis

List interfaces for devices in the inventory.

Examples:
  cani show interfaces                     # Show all interfaces
  cani show interfaces --format table      # Show as table
  cani show interfaces --format json       # Show as JSON
  cani show interfaces --device "server1"  # Show interfaces for specific device
  cani show interfaces --type mgmt         # Show only management interfaces

```
cani alpha show interfaces [flags]
```

### Options

```
      --device string   Filter interfaces by device name
  -h, --help            help for interfaces
      --type string     Filter interfaces by type (mgmt, ethernet, infiniband, sfp, osfp, outlet)
```

### Options inherited from parent commands

```
      --cable-type string     Filter cables by type (e.g., 'dac', 'cat6', used with --rack-view)
      --columns int           Number of rack columns before wrapping (0=auto, used with --rack-view)
      --config string         config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string      datastore type (json, postgres) (default "json")
      --debug                 enable debug mode
  -f, --file string           Load inventory from YAML file (used with --visual)
  -o, --format string         Output format (json) (default "json")
      --no-color              Disable colorized output (used with --visual)
      --rack string           Filter to specific rack by name (used with --visual)
      --rack-view             Display compact ASCII rack view with device symbols
      --show-cables           Show cable connections in visual output
      --show-routing          Display cable routing with branching visualization (1 rack per line)
  -s, --sort string           Sort by this field (name, type, id, status, vendor, model) (default "name")
      --strict                require a resolved device type (slug) for all devices (default true)
      --types-dirs strings    local directories with additional hardware types
      --types-repo-pull       pull latest changes from types repos on startup
      --types-repos strings   git repo URLs with additional hardware types
  -V, --verbose count         Verbose output: -V shows legend, -VV shows all cables (used with --rack-view)
  -v, --visual                Display ASCII visualization of rack layout
```

### SEE ALSO

* [cani alpha show](cani_alpha_show.md)	 - Show items from the inventory

