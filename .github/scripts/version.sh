#!/bin/sh

bump_semver_by_commit() {
  local version=$1
  local commit_message=$2

  # Default to patch (index 2)
  local index=2
  if echo "$commit_message" | grep -q '\[major\]'; then
    index=0
  elif echo "$commit_message" | grep -q '\[minor\]'; then
    index=1
  fi

  IFS='.' read -ra parts <<< "$version"
  parts[$index]=$((parts[$index] + 1))

  # Reset all lower-order parts
  for ((i = index + 1; i < ${#parts[@]}; i++)); do
    parts[$i]=0
  done

  echo "${parts[*]}" | tr ' ' '.'
}

bump_semver_by_commit "$1" "$2"