## cani init

Generate a new provider scaffold

### Synopsis

Generate a new provider scaffold with stubbed implementations.

This creates a new provider package with all required Provider interface methods
and optional interface stubs (Loader, HasOptions, HasImportOptions, HasExportOptions,
CableTransformer) with TODO comments.

Example:
  cani init mycloud
  cani init mycloud --output ./custom/path
  cani init mycloud --force  # Overwrite existing directory

```
cani init <provider-name> [flags]
```

### Options

```
  -f, --force           Overwrite existing directory
  -h, --help            help for init
  -o, --output string   Output directory (default: pkg/provider/<name>)
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

* [cani](cani.md)	 - Continious And Never-ending Inventory

