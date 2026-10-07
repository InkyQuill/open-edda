# Checks, versions and distribution

GitHub is the source/release host. GitHub Actions runs formatting, Go vet/race
checks, frontend typecheck/build/unit tests, headless desktop/mobile tests and the
real CLI terminal test. PR checks do not use deployment credentials or a
self-hosted runner.

`release-please` follows the Hieronymus workflow: Conventional Commits produce a
release PR changing `version.txt`, `.release-please-manifest.json` and
`CHANGELOG.md`. The initial unreleased baseline is `0.0.0`. Merge the release PR
when the version is ready. The workflow explicitly dispatches checks on the bot
branch because a PR created with `GITHUB_TOKEN` does not trigger normal PR events.

After checks pass on the merged version commit, an immutable `vX.Y.Z` tag and a
draft release are created. Native Linux amd64/arm64 runners build static musl CLI
archives; both builds must pass before the draft is published with `SHA256SUMS`.
`edda version` reports the embedded release version. Windows/macOS binaries are
not advertised: local checkout operations currently depend on Linux APIs.

If a release build fails, the draft remains private. Repair/rerun **Release CLI**
with that existing tag; it refuses to overwrite a published release. Source tags
are public even while release assets remain drafts. No new version is required
for a transient build failure. Enable “Allow GitHub Actions to create and approve
pull requests” in repository Actions settings. No personal access token is needed.

## Containers remain private

No workflow publishes an OCI image, uploads an image archive to GitHub, or logs
into a registry. Current deployment remains a local build, private archive
transfer and import into k3s (`imagePullPolicy: Never`). The `docker.io/library`
name in a local imported image does not imply a Docker Hub upload.

The next distribution step can use the existing **private GitLab Container
Registry**, independently of GitHub source/releases. Use a dedicated GitLab
project/package, a narrowly scoped publishing credential in a protected GitHub
environment, and a read-only image-pull credential in Kubernetes. Pin deployment
to the resulting image digest. This is a proposed option; no registry project,
credentials, public images or publication jobs have been created.
