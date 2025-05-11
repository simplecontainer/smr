#!/bin/bash

commit_message="$1"

matches=$(echo "$commit_message" | grep -o '\[[^][]\+\]')

for match in $matches; do
  echo "${match:1:-1}"
done
