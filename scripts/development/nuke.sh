#!/bin/bash

# Set image name (without tag)
IMAGE_NAME="smr"

# List all image tags for the given image
tags=$(docker images --format "{{.Repository}}:{{.Tag}} {{.CreatedAt}}" | grep "^$IMAGE_NAME:" | sort -rk2)

# Extract just the tag names (sorted by newest first)
tag_list=($(echo "$tags" | awk '{print $1}'))

# Keep first 2 tags
keep_tags=("${tag_list[@]:0:2}")

# Tags to delete
delete_tags=("${tag_list[@]:2}")

# Delete old tags
for tag in "${delete_tags[@]}"; do
    echo "Deleting tag: $tag"
    docker rmi "$tag"
done

echo "Kept tags:"
for tag in "${keep_tags[@]}"; do
    echo "  $tag"
done
