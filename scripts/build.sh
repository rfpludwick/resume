#!/usr/bin/env bash

set -exo pipefail

cd "$(dirname "$(dirname "$(realpath "${0}")")")"

go run .
