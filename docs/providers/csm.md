# CSM Provider

The CSM provider imports from and exports to SLS (System Layout Service), CSM's hardware inventory backend.

## Import

```shell
# Import from CSM on an NCN
cani alpha import csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host api-gw-service-nmn.local

# Import from CSM over the CMN
cani alpha import csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host cmn.example.com
```

## Export

```shell
# Export to CSM
cani alpha export csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host api-gw-service-nmn.local
```

## Configuration

Provider-specific options in `~/.cani/cani.yml`:

```yaml
providers:
  csm:
    keycloak_username: ""       # Keycloak username
    keycloak_password: ""       # Keycloak password
    api_host: ""                # API gateway host
    ca_cert: ""                 # Path to CA certificate (optional)
    insecure: false             # Skip TLS verification (not recommended)
```
