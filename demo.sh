#!/usr/bin/env bash
# Run from ncn-m002!

if [[ $DEBUG == "true" ]]; then
  set -x
fi

# show auth using system certs
rm -rf ~/.cani && bin/cani alpha session start csm \
  --csm-keycloak-username vshasta \
  --csm-keycloak-password vshasta \
  --csm-base-auth-url https://api-gw-service-nmn.local/ \
  --csm-url-sls https://api-gw-service-nmn.local/apis/sls/v1

# show only one system in the inventory added by default (also show the basic inventory structure)
bin/cani alpha list | jq

# add a new cabinet
bin/cani alpha add cabinet
bin/cani alpha add cabinet hpe-ex2000
# show various errors
bin/cani alpha add cabinet hpe-ex2000 --cabinet 1
bin/cani alpha add cabinet hpe-ex2000 --vlan-id 1
# show successful add
bin/cani alpha add cabinet hpe-ex2000 --vlan-id 1 --cabinet 1
bin/cani alpha list cabinet | jq
# optionally, show auto-added child hardware
bin/cani alpha list | jq
# show where it comes from
ls -l pkg/hardwaretypes/hardware-types
# and how it is defined
head pkg/hardwaretypes/hardware-types/hpe-cabinet-ex2000.yaml
# folder for user-defined configs TODO

# try to add the same cabinet number should fail
bin/cani alpha add cabinet hpe-ex2000 --vlan-id 1 --cabinet 1
# same for the same vlan-id (this does not work currently since we aren't checking vlan uniqueness)
# bin/cani alpha add cabinet hpe-ex2000 --vlan-id 1 --cabinet 2

# add blade in the in same way
bin/cani alpha add blade hpe-crayex-ex420-compute-blade --cabinet 1 --chassis 1 --blade 0
bin/cani alpha update node --cabinet 1 --chassis 1 --blade 0 --nodecard 1 --node 0 --role "Compute" --alias "nid00001" --nid 1

# stop the session and validate/commit changes; push to SLS
bin/cani alpha session stop
