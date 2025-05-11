#!/bin/bash

bump_semver_by_commit() {
  local version="$1"
  local commit_message="$2"

  local index=2
  if echo "$commit_message" | grep -q '\[major\]'; then
    index=0
  elif echo "$commit_message" | grep -q '\[minor\]'; then
    index=1
  fi

  IFS='.' read -r -a parts <<< "$version"
  parts[$index]=$((parts[$index] + 1))

  for ((i = index + 1; i < ${#parts[@]}; i++)); do
    parts[$i]=0
  done

  (IFS=.; echo "${parts[*]}")
}

bump_semver_by_commit "$1" "$2"