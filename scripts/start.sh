#!/bin/sh
set -e

go build -o /app/bin/api /app/src/*.go

/app/bin/api
