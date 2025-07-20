#!/bin/bash

set -e

# Build the server binary
go build -o server cmd/server.go

# Run the server
./server --port $1
