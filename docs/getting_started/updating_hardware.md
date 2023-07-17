# Updating Hardware

Some providers and some certain types of hardware require additional metadata.

## Update Hardware

CSM is one such entity.  NodeBlades require additional metadata to work properly in SLS and HSM.  Running `validate` will report any such issues and the `update` subcommand can be used to make the appropriate changes.

```shell
# add a cabinet
cani alpha add cabinet hpe-ex2000 --auto --accept
# add a blade, which also adds a node or nodes
cani alpha add blade hpe-crayex-ex420-compute-blade --cabinet 9000 --chassis 0 --blade 0
cani alpha validate # shows provider-specific errors such as missing roles or aliases
# update the node or nodes
cani alpha update node --cabinet 3000 --chassis 0 --blade 0 --node-card 1 --node 0 --role "Compute" --alias "nid00001" --nid 1
cani alpha validate # validates ok now
```



