#!/bin/bash

export API_HOST="$(uname -n):8080"

if [[ "$1" == "mainstay" ]]; then
    echo "Running attestation"
    mainstay -tx ${TX}
else
  $@
fi
