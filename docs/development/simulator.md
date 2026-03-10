# Local Development Using Simulators

Developing locally is a quick way to iterate. Both the CSM and Nautobot simulators
run via Docker Compose and have corresponding Makefile targets.

## CSM Simulator

### Start The Simulator

The CSM simulator runs an SLS instance behind an NGINX API gateway. Self-signed
TLS certificates are generated automatically.

```shell
# Start the simulator (generates certs, starts containers, loads SLS data)
make csm-up

# Optionally load a different SLS fixture
make csm-up SLS_FILE=testdata/fixtures/csm/sls/valid-mug.json
```

### Import From CSM

> The `-S` flag enables simulator mode and `-k` bypasses certificate warnings.
> OAuth is bypassed for the simulator.

```shell
# Import from the simulator
rm -rf ~/.cani && go run . alpha import csm -S -k

# SLS is pre-populated, but it is often useful to empty it or populate it with test data
# Wipe out SLS
echo '{}' > /tmp/sls_empty.json; curl -X POST -F "sls_dump=@/tmp/sls_empty.json" https://localhost:8443/apis/sls/v1/loadstate -ik

# Load a known-good config
curl -X POST -F "sls_dump=@testdata/fixtures/csm/sls/valid-mug.json" https://localhost:8443/apis/sls/v1/loadstate -ik

# Run cani commands against the local data
go run . alpha show device
go run . alpha add device hpe-crayex-ex420-compute-blade --auto --accept

# Export back to the simulator
go run . alpha export csm -S -k
```

### Stop The CSM Simulator

```shell
make csm-down
```

### vShasta v2

Testing `cani` on vShasta v2 provides a useful development environment with simulated SLS, HSM, and Redfish services. Since there is no real hardware, it is safe to interact with these services.

#### Prerequisites

1. Install the [gcloud CLI](https://cloud.google.com/sdk/docs/install)
2. Request an algol60 account and a vShasta v2 instance via the `#vshasta2` Slack channel

#### Usage

```shell
# Authenticate and generate SSH config
gcloud auth login
gcloud compute config-ssh --project project-id

# SSH to a node
ssh ncn-m002.us-central1-a.project-id

# Install Go: https://go.dev/doc/install
# Clone the repo
git clone https://github.com/Cray-HPE/cani.git && cd cani

# Import from CSM on vShasta
go run . alpha import csm \
  --csm-keycloak-username vshasta \
  --csm-keycloak-password vshasta \
  --csm-api-host api-gw-service-nmn.local

# Work with the imported data
go run . alpha show device
go run . alpha classify --auto

# Export back to CSM
go run . alpha export csm \
  --csm-keycloak-username vshasta \
  --csm-keycloak-password vshasta \
  --csm-api-host api-gw-service-nmn.local
```

> There is no CMN on vShasta v2, so external access to the instance is not currently possible. If running from the PIT node, copy the CA cert from another node and pass `--csm-ca-cert`.

## Nautobot Simulator

### Start Nautobot

The Nautobot simulator runs Nautobot with a PostgreSQL backend via Docker Compose.

```shell
# Start Nautobot (defaults to latest version, Python 3.11)
make nautobot-up

# Start a specific Nautobot version
make nautobot-up NAUTOBOT_VERSION=2.0.0 PYTHON_VER=3.11
```

Nautobot will be available at `http://localhost:8081`. 

### Import From Nautobot

```shell
# Import from the local Nautobot instance
go run . alpha import nautobot

# Run cani commands against the local data
go run . alpha show device
go run . alpha add rack --auto --accept

# Export back to Nautobot
go run . alpha export nautobot
```

### Useful Nautobot Commands

```shell
make nautobot-logs    # Tail service logs
make nautobot-shell   # Open a shell in the Nautobot container
```

### Stop Nautobot

```shell
make nautobot-down
```
