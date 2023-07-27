#! /usr/bin/env bash
#
# MIT License
#
# (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
set -eux

SCRIPT_DIR=$(dirname "$0")
echo "Script directory: ${SCRIPT_DIR}"

source "${SCRIPT_DIR}/../../venv/bin/activate"

if [[ -z "$1" ]]; then
    echo "Missing system name argument!"
    echo "usage: ${0} system_xname "
fi
SYSTEM_NAME="$1"
echo "System name: ${SYSTEM_NAME}"

# Load SLS
SLS_FILE="${SCRIPT_DIR}/${SYSTEM_NAME}/sls_dumpstate.json"
curl -k -X POST -F "sls_dump=@${SLS_FILE}" https://localhost:8443/apis/sls/v1/loadstate -i
"${SCRIPT_DIR}/adding_missing_mtn_networks.sh"

# Load HSM
FILL_SMD_DB_FULLPATH=$(realpath "${SCRIPT_DIR}/fill-smd-db.sh")
SMD_SQL_PATH=$(realpath "${SCRIPT_DIR}/${SYSTEM_NAME}/smd.sql")
pushd "${SCRIPT_DIR}/../../hms-simulation-environment" || exit
    "${FILL_SMD_DB_FULLPATH}" "${SMD_SQL_PATH}"
popd

# Load BSS
"${SCRIPT_DIR}/load_bss.py" "${SCRIPT_DIR}/${SYSTEM_NAME}/bss_bootparameters.json"

