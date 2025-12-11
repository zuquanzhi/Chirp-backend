#!/bin/bash
set -euo pipefail

CONFIG_FILE=${CONFIG_FILE:-config.json}

echo "Using config file: ${CONFIG_FILE}"
if [[ ! -f "${CONFIG_FILE}" ]]; then
	echo "Config file not found: ${CONFIG_FILE}. Copy config.example.json and modify as needed."
	exit 1
fi

echo "Checking dependencies..."
go mod tidy

echo "Starting Chirp Server... (CONFIG_FILE=${CONFIG_FILE})"
echo "Press Ctrl+C to stop."
CONFIG_FILE="${CONFIG_FILE}" go run ./cmd/server/main.go
