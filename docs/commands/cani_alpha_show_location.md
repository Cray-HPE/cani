## cani alpha show location

List locations in the inventory.

### Synopsis

List locations in the inventory.

```
cani alpha show location [flags]
```

### Options

```
  -h, --help   help for location
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

