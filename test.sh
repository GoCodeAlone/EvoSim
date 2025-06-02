#!/bin/bash

# This script runs tests without being affected by go.work files
echo "Running tests (isolated from go.work)..."
GOWORK=off go test -v
