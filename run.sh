#!/bin/bash

# This script runs the genetic algorithm without being affected by go.work files
echo "Running Genetic Algorithm (isolated from go.work)..."
GOWORK=off go run .
