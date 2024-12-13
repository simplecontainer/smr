#!/bin/bash
cd "$(dirname "$0")"

echo "Doing work in directory $PWD"

ENCRYPTED=$(../production/smrmgr.sh export https://localhost:1443)
KEY=$(cat "$HOME/smr/smr/contexts/$(smr context).key")

../production/smrmgr.sh import $ENCRYPTED <<< $KEY