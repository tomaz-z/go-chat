#!/bin/sh
set -e

cd /app/src

go build -o /app/bin/api /app/src/cmd/server

exec /app/bin/api
