#!/usr/bin/env bash
#
# MIT License
#
# (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
set -e
set -u
set -o pipefail

SIM_REPO="${1:-../hms-simulation-environment}"
SLS_DUMP="${2:-../hms-simulation-environment/configs/sls/no_hardware.json}"
CANI_COMPOSE="${3:-./testdata/docker-compose.cani.yml}"

# start the simulator with the specified SLS config
pushd "${SIM_REPO}" || exit 1
  ./setup_venv.sh
  #shellcheck disable=SC1091
  . ./venv/bin/activate
  ./run.py "${SLS_DUMP}"

  # Stop Discovery services as they are not currently required.
  # Some edge case tests cause high load with HSM and MEDS as they are 
  # trying to reach out to hardware that does not exist.
  docker compose stop cray-meds
  docker compose stop cray-reds
popd || exit 1

# # setup additonal simulation services for other providers
# docker compose -f "${CANI_COMPOSE}" up -d
# # export token=$(docker exec testdata-cmdb-1 cat /opt/clmgr/etc/.cli_token) 
# docker volume create --name cmdb-data --opt type=none --opt device=testdata/fixtures/hpcm/
