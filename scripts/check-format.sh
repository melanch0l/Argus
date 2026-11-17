#!/bin/sh
set -eu
files=$(gofmt -l cmd internal)
if [ -n "$files" ]; then
    printf 'Run make fmt on:\n%s\n' "$files" >&2
    exit 1
fi
