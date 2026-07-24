#!/bin/bash
# Checks the REPO-MAP block in CLAUDE.md against the actual package tree.
# Fails when a package directory exists that the map doesn't mention, so the
# index agents rely on can't silently drift. Update the block, not this script.
set -euo pipefail
cd "$(dirname "$0")/.."

missing=0
while read -r d; do
    case "$d" in
        # Covered by a parent or wildcard line in the map.
        internal/pkg/helper/*) d='internal/pkg/helper/\*' ;;
        internal/application/*/*) continue ;;
        internal/server/api/docs) continue ;;      # generated swagger
        internal/server/grpc/proto*) continue ;;   # generated stubs
        cmd/migration/*) continue ;;
        internal/common/type/*) continue ;;
        cmd|internal|internal/common|internal/pkg|internal/application|internal/server) continue ;;
    esac
    if ! grep -q "$d" CLAUDE.md; then
        echo "REPO-MAP missing: $d"
        missing=1
    fi
done < <(find cmd infrastructure internal -type d | sort -u)

if [ "$missing" -ne 0 ]; then
    echo "CLAUDE.md REPO-MAP is stale — add the packages above to the block."
    exit 1
fi
echo "REPO-MAP up to date."
