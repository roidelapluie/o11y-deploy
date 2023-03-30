#!/usr/bin/env bash

# check if at least two arguments are provided
if [ $# -lt 2 ]; then
  echo "Usage: $0 <tar_file> <directory1> [<directory2> ...]"
  exit 1
fi

# assign the first argument as the tar.gz file
tar_file="$1"
shift

# create a temporary directory to hold the directories
temp_dir=$(mktemp -d)

# copy each directory to the temp directory
for directory in "$@"; do
  cp -r "$directory" "$temp_dir/$(basename "$directory")"
done

# create the tar.gz archive
tar -czf "$tar_file" -C "$temp_dir" .

# remove the temporary directory
rm -rf "$temp_dir"
