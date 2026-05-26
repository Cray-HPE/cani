## cani alpha export csm

Export inventory using the csm provider

```
cani alpha export csm [flags]
```

### Options

```
  -a, --all                            Include all hardware types (overrides --type)
      --commit                         Push changes to the remote CSM system
      --csm-api-host string            FQDN or host:port of the CSM API gateway (default "api-gw-service.local")
      --csm-ca-cert string             Path to a PEM-encoded CA certificate
      --csm-k8s-pods-cidr string       CIDR used by Kubernetes for pods (default "10.32.0.0/12")
      --csm-k8s-services-cidr string   CIDR used by Kubernetes for services (default "10.16.0.0/12")
      --csm-keycloak-password string   Keycloak password for authentication
      --csm-keycloak-username string   Keycloak username for authentication
      --csm-url-hsm string             Override the HSM API base URL
      --csm-url-sls string             Override the SLS API base URL
      --dryrun                         Perform a dry run without applying changes
      --format string                  Output format: csv, sls-json (default "csv")
      --headers string                 Comma-separated list of CSV columns to include (default "Type,Vlan,Role,SubRole,Status,Nid,Alias,Name,ID")
  -h, --help                           help for csm
      --ignore-validation              Skip validation (only applies to sls-json format)
  -k, --insecure                       Skip TLS certificate verification
  -t, --type string                    Comma-separated list of hardware types to include (default "Node,Cabinet")
  -S, --use-simulator                  Use simulation mode (localhost:8443, no auth, skip TLS verification)
```

### Options inherited from parent commands

```
      --config string         config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string      datastore type (json, postgres) (default "json")
      --debug                 enable debug mode
      --dry-run               Preview changes without making API calls
      --merge                 Update existing devices instead of skipping conflicts
      --strict                require a resolved device type (slug) for all devices (default true)
      --types-dirs strings    local directories with additional hardware types
      --types-repo-pull       pull latest changes from types repos on startup
      --types-repos strings   git repo URLs with additional hardware types
```

### SEE ALSO

* [cani alpha export](cani_alpha_export.md)	 - Export inventory to an external provider

