# Local Development Using The HMS Simulator

Developing locally is a quick way to iterate.

## Setup The HMS Simulator

```shell
git clone https://github.com/Cray-HPE/hms-simulation-environment.git
./setup_venv.sh
source ./venv/bin/activate
./run.py configs/sls/small_mountain.json # or another configuration
```

## Start A `cani` Session Using The Simulator

> Built into cani, is support for the simulator, which is tied to a single flag, -S
> to by pass certificate warnings, also pass -k
> Oauth is bypassed for the simulator

```shell
# start a session using the --simulation flag
rm -rf ~/.cani && go run . alpha session start csm -S -k
# sls is pre-populated, but it is often useful to empty it out or populate it with something else
# wipe out SLS
echo '{}' > /tmp/sls_empty.json; curl -X POST -F "sls_dump=@/tmp/sls_empty.json" https://localhost:8443/apis/sls/v1/loadstate -ik
# put in a known good-config (this could be any valid SLS file)
curl -X POST -F "sls_dump=@testdata/fixtures/sls/mug-dumpstate.json" https://localhost:8443/apis/sls/v1/loadstate -ik
# run cani commands...
```
