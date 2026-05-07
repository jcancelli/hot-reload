#!/usr/bin/env bash

set -euo pipefail

PROJECT_DIR="$(realpath $(dirname "${BASH_SOURCE[0]}"))"

# Removes older version
go clean -i $PROJECT_DIR

# Installs newer version
go install $PROJECT_DIR
