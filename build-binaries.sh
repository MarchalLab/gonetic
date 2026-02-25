#!/usr/bin/env bash

set -e

OUTDIR="gonetic-binaries"
APPNAME="gonetic"

echo "Creating output directory: ${OUTDIR}"
rm -rf "${OUTDIR}"
mkdir -p "${OUTDIR}"

echo "Building Linux (amd64)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o "${OUTDIR}/${APPNAME}-linux-amd64"

echo "Building Windows (amd64)..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
    go build -o "${OUTDIR}/${APPNAME}-windows-amd64.exe"

echo "Building macOS (Intel amd64)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
    go build -o "${OUTDIR}/${APPNAME}-darwin-amd64"

echo "Building macOS (Apple Silicon arm64)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
    go build -o "${OUTDIR}/${APPNAME}-darwin-arm64"

echo "Build complete. Binaries are in ${OUTDIR}/"