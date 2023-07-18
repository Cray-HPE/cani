#! /usr/bin/env bash

set -eux

# TODO write down collect steps

SCRIPT_DIR=$(dirname "$0")
echo "Script directory: ${SCRIPT_DIR}"

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

