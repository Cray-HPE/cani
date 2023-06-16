#!/usr/bin/env bash
# Run from ncn-m002!

if [[ $DEBUG == "true" ]]; then
  set -x
fi

# wipe SLS
# echo '{}' > /tmp/sls_empty.json; curl -X POST -F "sls_dump=@/tmp/sls_empty.json" https://api-gw-service-nmn.local/apis/sls/v1/loadstate -ik
# show empty
# curl -k https://localhost:8443/apis/sls/v1/dumpstate | jq
# populate SLS with a known good config
# curl -X POST -F "sls_dump=@../sls_dump_1686855722.json" https://api-gw-service-nmn.local/apis/sls/v1/loadstate -ik

# show auth using system certs
rm -rf ~/.cani && bin/cani alpha session start csm \
  --csm-keycloak-username vshasta \
  --csm-keycloak-password vshasta \
  --csm-base-auth-url https://api-gw-service-nmn.local/ 

# Show what SLS looks like before the import for a piece of hardware 
cray sls hardware describe x1000c0s3b0n0 --format json | jq

# Perform the import of CSM data from SLS. 
bin/cani alpha session import 

# Show what changed in SLS for the node
cray sls hardware describe x1000c0s3b0n0 --format json | jq

# The @cani.id property has the ID of the inventory object that created it
# NEED TO UPDATE THE UUID TO MATCH
cat ~/.cani/canidb.json | jq '.Hardware."26251bc3-9a2c-4658-a90d-d897775c8ba0"'

# Run again to show idempotencey
bin/cani alpha session import



# show only one system in the inventory added by default (also show the basic inventory structure)
bin/cani alpha list | jq

# add a new cabinet
bin/cani alpha add cabinet
bin/cani alpha add cabinet hpe-ex4000
# show various errors
bin/cani alpha add cabinet hpe-ex4000 --cabinet 1005
bin/cani alpha add cabinet hpe-ex4000 --vlan-id 4000
# show successful add
bin/cani alpha add cabinet hpe-ex4000 --vlan-id 4000 --cabinet 1005
bin/cani alpha list cabinet | jq
# optionally, show auto-added child hardware
bin/cani alpha list | jq
# show where it comes from
ls -l pkg/hardwaretypes/hardware-types
# and how it is defined
head pkg/hardwaretypes/hardware-types/hpe-cabinet-ex4000.yaml
# folder for user-defined configs TODO

# try to add the same cabinet number should fail
bin/cani alpha add cabinet hpe-ex2000 --vlan-id 4000 --cabinet 1005
# same for the same vlan-id (this does not work currently since we aren't checking vlan uniqueness)
# TODO TEST again, new changes might make this work
bin/cani alpha add cabinet hpe-ex2000 --vlan-id 4000 --cabinet 2

# add blade in the in same way
bin/cani alpha add blade hpe-crayex-ex235a-compute-blade --cabinet 1005 --chassis 1 --blade 1

# Show that what is missing
bin/cani alpha validate
bin/cani alpha update node --cabinet 1005 --chassis 1 --blade 1 --nodecard 0 --node 0 \
    --role Compute --nid 3000 --alias nid003000

bin/cani alpha validate

# Show that duplicate/bad data can't be put in
# Invalid Role
bin/cani alpha update node --cabinet 1005 --chassis 1 --blade 1 --nodecard 1 --node 0 \
    --role MyCompute --nid 3001 --alias nid003001
# Duplicate NID and alias
bin/cani alpha update node --cabinet 1005 --chassis 1 --blade 1 --nodecard 1 --node 0 \
    --role Compute --nid 3000 --alias nid003000

# Update the second node on teh blade for real
bin/cani alpha update node --cabinet 1005 --chassis 1 --blade 1 --nodecard 1 --node 0 \
    --role Compute --nid 3001 --alias nid003001

# Show no issues
bin/cani alpha validate

# stop the session and validate/commit changes; push to SLS
bin/cani alpha session stop
bin/cani alpha session stop # Run again to show idempotencey

# Show the new hardware in SLS!!!
cray sls hardware describe x1005c1s1b0n0 --format json
cray sls hardware describe x1005c1s1b1n0 --format json
cray sls hardware describe x1005c1 --format json
cray sls hardware describe x1005c1b0 --format json

# show auth against a real machine using CMN
rm -rf ~/.cani && CANI_USER="" CANI_PASSWORD="" bin/cani alpha session start csm \
 --csm-keycloak-username "$CANI_USER" \
 --csm-keycloak-password "$CANI_PASSWORD" \
 --csm-base-auth-url https://auth.cmn.drax.hpc.amslabs.hpecorp.net \
 --csm-url-sls https://api.cmn.drax.hpc.amslabs.hpecorp.net/apis/sls/v1 \
 -k
