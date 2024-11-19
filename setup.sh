#!/bin/bash

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # List of required packages
    declare -a extra_packages=(
        "build-essential"
    )
    # Linux: Use dpkg to check for packages
    for package in "${extra_packages[@]}"; do
        if ! dpkg -l | grep -q "^ii\s*$package\s"; then
            printf "Package %s is missing on Linux\n" "$package"
            exit 1
        fi
    done
fi
