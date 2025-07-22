#!/bin/bash

set -e

protoc --go_out=. --go-grpc_out=. ./proto/dumbkv.proto