#!/usr/bin/env bash

set -e

path=${1:?Please provide photo library root as first argument}
export PHOTO_ROOT="${path}"
go run .