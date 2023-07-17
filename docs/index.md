# CANI

`cani` is a hardware inventory tool.  It provides its own portable inventory format, while retaining compatiblity with external inventory providers.  This makes it possible to use `cani` as either a main inventory source or to migrate from one inventory format to another.  

## Quickstart

This shows a quick overview of using `cani` to connect to an external inventory provider, add/update hardware in `cani`, and then commit the changes back to the provider.

```shell
# start a session with a provider (CSM in this case)
cani alpha session start csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host api-gw-service-nmn.local

# import data from the provider
cani alpha session import

# add a cabinet, accepting recommended values
cani alpha add cabinet hpe-ex2000 --auto --accept 

# add a blade, which also adds a node or nodes
cani alpha add blade hpe-crayex-ex420-compute-blade --cabinet 9000 --chassis 0 --blade 0

# validate the data at any time
cani alpha validate # shows provider-specific errors such as missing roles or aliases

# update the node or nodes
cani alpha update node --cabinet 3000 --chassis 0 --blade 0 --node-card 1 --node 0 --role "Compute" --alias "nid00001" --nid 1

# stop the session and commit the data to the external inventory provider (CSM's SLS in this example)
cani alpha session stop
```

