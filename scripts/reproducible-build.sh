#!/bin/sh
set -e

go build -trimpath -o /tmp/replaynet-build1 ./cmd/replaynet
go build -trimpath -o /tmp/replaynet-build2 ./cmd/replaynet

HASH1=$(sha256sum /tmp/replaynet-build1 | cut -d' ' -f1)
HASH2=$(sha256sum /tmp/replaynet-build2 | cut -d' ' -f1)

echo "build 1: $HASH1"
echo "build 2: $HASH2"

if [ "$HASH1" = "$HASH2" ]; then
  echo "MATCH: reproducible build confirmed"
else
  echo "MISMATCH: builds are not identical"
  exit 1
fi
