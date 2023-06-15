# Testing On vShasta v2

Testing `cani` on Vshasta2 has some prerequisite steps, but can provide a useful development environment.  Since there is not actually any hardware on Vshasta2, it is fairly safe to interact with SLS, HSM, and the simulated redfish environment.

## Prerequisites

### Gcloud CLI

Install gcloud: https://cloud.google.com/sdk/docs/install

### Get Access To A sVhasta v2 Instance

1. Request an algol60 account
2. Request a vShasta v2

Both of these are available via the `#vshasta2` Slack channel

### Log Onto The Instance

```shell
# authenticate
gcloud auth login
# generate an ssh config to use ssh directly instead of gcloud ssh
gcloud compute config-ssh --project project-id # see also: gcloud compute config-ssh --remove
# ssh to the pit or another node
ssh ncn-m002.us-central1-a.project-id # see also: gcloud ssh --project project-id ncn-m002
# install go on ncn-m002: https://go.dev/doc/install
# clone cani repo
git clone https://github.com/Cray-HPE/cani.git && cd cani
# start a cani session on vshasta
go run . alpha session start csm \
--csm-keycloak-username vshasta \
--csm-keycloak-password vshasta \
--csm-base-auth-url=https://api-gw-service-nmn.local
# if running from the pit, copy the cert from another node and pass --csm-ca-cert
```

> There is no CMN on vShasta v2 so it is not currently possible to query a vShasta instance from outside the nodes themselves.
