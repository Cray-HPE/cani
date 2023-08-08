# Starting A Session

All `cani` operations fall under the purview of a __session__ in which hardware can be added, removed, or updated and is not commited to the external inventory provider until the session is stopped.

## Start A Session With A Provider

To start a session, you need to choose an inventory provider to work with.  At present, the only available option is `csm`.  During session setup, it is also necessary to provide credentials and paths to resources like SLS and HSM.

A common way to interact with a provider is from a local laptop.  With the CSM provider, the services are often available over the CMN.

Starting a simple session might look like this, which authorizes a user over the CMN (auth.cmn.example.com and api.cmn.example.com).

> Note: This example uses -k (insecure), which is not recommended, but is available for this alpha release

```shell
# example: starting a session over the CMN
cani alpha session init csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host cmn.example.com \
  -k
```

Just as common, is running `cani` on an NCN.

> Note: since the NCNs have the platform certificates installed, the -k flag is not used

```shell
# example: starting a session on an NCN
cani alpha session init csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host api-gw-service-nmn.local
```
