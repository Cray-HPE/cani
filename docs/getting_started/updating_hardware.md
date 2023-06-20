# Updating Hardware

Some providers and some certain types of hardware require additional metadata.

## Update Hardware

CSM is one such entity.  NodeBlades require additional metadata to work properly in SLS and HSM.  Running `validate` will report any such issues and the `update` subcommand can be used to make the appropriate changes.

```shell
# example setting a HMN vlan for a new cabinet
cani alpha add cabinet ex2000
cani alpha validate # shows an error about a missing vlan 
cani alpha update cabinet --vlan-id 4
cani alpha validate # validates ok now
```



