#!/usr/bin/env bash
# mkdir -p cani_demo && cd !$
# git clone https://github.com/Cray-HPE/cani.git 
# git checkout supreme-sacrifice

pushd ~/cani_demo/cani || exit 1
  # ensure go is in the path for this script
  export PATH=$PATH:/usr/local/go/bin
  if ! command -v go; then
    # get and install go
    wget https://go.dev/dl/go1.20.5.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.5.linux-amd64.tar.gz
  fi
  
  # modify SLS to pass checks
  # cray sls dumpstate list --format json > sls_dump_"$(date +"%s")".json # Backup just in case :)
  # cray sls networks update ./hmn_mtn.json HMN_MTN
  # cray sls networks update ./nmn_mtn.json NMN_MTN

  # Make the binary
  make bin

popd || exit 1

