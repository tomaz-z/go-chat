#!/bin/sh
set -e

cd /app/src/server

go build -o /app/bin/api /app/src/server

/app/bin/api
