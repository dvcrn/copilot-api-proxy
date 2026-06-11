#!/bin/bash
cd "$(dirname "$0")" && go run ./cmd/copilot-oauth-proxy/main.go server
