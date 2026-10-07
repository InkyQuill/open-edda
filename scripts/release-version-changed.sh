#!/usr/bin/env bash
# Caller checks out full history. Missing history must fail, not skip a release.
set -euo pipefail
if [ "${GITHUB_EVENT_NAME:?}" != push ] || [[ "${BEFORE:?}" =~ ^0+$ ]]; then
  echo false
  exit 0
fi
git cat-file -e "$BEFORE^{commit}"
if git cat-file -e "$BEFORE:version.txt" 2>/dev/null; then
  previous=$(git show "$BEFORE:version.txt")
else
  # First installation of release automation has no previous version file.
  previous=0.0.0
fi
if [ "$(cat version.txt)" = "$previous" ]; then
  echo false
else
  echo true
fi
