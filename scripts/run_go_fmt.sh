#!/bin/bash

set -eu
set -o pipefail

golangci-lint fmt --stdin
