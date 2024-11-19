#!/bin/bash

# List of required packages
declare -a extra_packages=(
    "build-essential"
)

# Check for each package and add missing ones to the list
for package in "${extra_packages[@]}"; do
    if ! dpkg -l | grep -q "^ii\s*$package\s"; then
        printf "Package %s is missing\n" "$package"
        return 1
    fi
done
