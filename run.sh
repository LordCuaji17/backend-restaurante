#!/usr/bin/env bash
# Arranca el servidor (lee .env automáticamente)
cd "$(dirname "$0")"
exec go run .
