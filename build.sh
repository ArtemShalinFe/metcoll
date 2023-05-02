#!/bin/bash
go build -buildvcs=false -C ./cmd/agent -o agent
go build -buildvcs=false -C ./cmd/server -o server
