#!/usr/bin/env bash
# Run inside the directory containing both CLI archives.
set -euo pipefail
[[ "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]
if [ "$(gh release view "$TAG" --json isDraft --jq .isDraft)" != true ]; then
  echo 'Release is already published; refusing to replace assets.' >&2
  exit 1
fi
test -f "edda-${TAG}-linux-amd64.tar.gz"
test -f "edda-${TAG}-linux-arm64.tar.gz"
sha256sum ./*.tar.gz > SHA256SUMS
gh release upload "$TAG" ./*.tar.gz SHA256SUMS --clobber
gh label create 'autorelease: tagged' --color 0E8A16 --force
numbers=$(gh api "repos/$GH_REPO/commits/$TAG/pulls" --jq '.[] | select(.merged_at != null and .base.ref == "main" and (.head.ref | startswith("release-please--branches--main"))) | .number')
for number in $numbers; do
  gh pr edit "$number" --add-label 'autorelease: tagged'
  pending=$(gh api "repos/$GH_REPO/issues/$number/labels" --jq '[.[] | select(.name == "autorelease: pending")] | length')
  if [ "$pending" != 0 ]; then
    gh pr edit "$number" --remove-label 'autorelease: pending'
  fi
done
gh release edit "$TAG" --draft=false
