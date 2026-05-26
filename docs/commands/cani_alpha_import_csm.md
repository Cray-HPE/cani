## cani alpha import csm

Import assets using the csm provider

```
cani alpha import csm [flags]
```

### Options

```
      --csm-api-host string            FQDN or host:port of the CSM API gateway (default "api-gw-service.local")
      --csm-ca-cert string             Path to a PEM-encoded CA certificate
      --csm-k8s-pods-cidr string       CIDR used by Kubernetes for pods (default "10.32.0.0/12")
      --csm-k8s-services-cidr string   CIDR used by Kubernetes for services (default "10.16.0.0/12")
      --csm-keycloak-password string   Keycloak password for authentication
      --csm-keycloak-username string   Keycloak username for authentication
      --csm-url-hsm string             Override the HSM API base URL
      --csm-url-sls string             Override the SLS API base URL
  -h, --help                           help for csm
      --ignore-validation              Continue importing even if the external inventory fails validation
  -k, --insecure                       Skip TLS certificate verification
      --sls-file string                Path to SLS dumpstate JSON file
      --smd-file string                Path to SMD state components JSON file
  -S, --use-simulator                  Use simulation mode (localhost:8443, no auth, skip TLS verification)
```

### Options inherited from parent commands

```
      --config string         config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string      datastore type (json, postgres) (default "json")
      --debug                 enable debug mode
      --no-color              Disable colorized output
      --phase string          ETL phases to run: e (extract), et (+transform), etl (+load) (default "etl")
      --step                  Step through each item interactively (implies --debug)
      --strict                require a resolved device type (slug) for all devices (default true)
      --types-dirs strings    local directories with additional hardware types
      --types-repo-pull       pull latest changes from types repos on startup
      --types-repos strings   git repo URLs with additional hardware types
```

### SEE ALSO

* [cani alpha import](cani_alpha_import.md)	 - Import assets into the inventory

