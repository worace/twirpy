#!/usr/bin/env bash

GOARCH=amd64 GOOS=linux go build . && cp protoc-gen-twirpy ~/code/medallion/docker/codegen/protoc-gen-twirpy-amd64
GOARCH=arm64 GOOS=linux go build . && cp protoc-gen-twirpy ~/code/medallion/docker/codegen/protoc-gen-twirpy-arm64
