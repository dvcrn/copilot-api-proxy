#!/bin/bash
cd "$(dirname "$0")" && go run ./cmd/copilot-api-proxy/main.go server
