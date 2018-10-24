#!/bin/bash

export API_HOST="$(uname -n):8080"

if [[ "$1" == "ocean-attestation" ]]; then
    echo "Running attestation"
    ocean-attestation -tx ${TX}
else
  $@
fi
