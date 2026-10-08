# Checks, versions and distribution

GitHub is the source/release host. GitHub Actions runs formatting, Go vet/race
checks, frontend typecheck/build/unit tests, headless desktop/mobile tests and the
real CLI terminal test. PR checks do not use deployment credentials or a
self-hosted runner.

`release-please` follows the Hieronymus workflow: Conventional Commits produce a
release PR changing `version.txt`, `.release-please-manifest.json` and
`CHANGELOG.md`. The initial unreleased baseline is `0.0.0`. Merge the release PR
when the version is ready. Bot PRs use the same `pull_request` checks as human
PRs. GitHub puts PR runs created/updated by `GITHUB_TOKEN` into an approval-required
state: a repository writer selects **Approve workflows to run** in the PR.
A manual `workflow_dispatch` run is not a substitute for required PR checks.
See [GitHub's trigger rules](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/trigger-a-workflow).

Release runs are serialized with `queue: max`, so newer pushes do not replace an
already pending version-change run (GitHub's queue limit is 100). Version detection
uses full Git history, including the commit before a multi-commit push. Missing
history fails explicitly. First adoption without an older version file uses the
unreleased `0.0.0` baseline.

After checks pass on the merged version commit, an immutable `vX.Y.Z` tag and a
draft release are created. Native Linux amd64/arm64 runners build static musl CLI
archives; both builds must pass before the draft is published with `SHA256SUMS`.
`edda version` reports the embedded release version. Windows/macOS binaries are
not advertised: local checkout operations currently depend on Linux APIs.

If a release build fails, the draft remains private. Repair/rerun **Release CLI**
with that existing tag; it refuses to overwrite a published release. Source tags
are public even while release assets remain drafts. No new version is required
for a transient build failure. Asset upload and retryable PR-label updates happen
while the release is still a draft; publishing is the final mutation. Metadata
failures therefore leave a draft that can be retried. Enable “Allow GitHub Actions to create and approve
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

## Local verification

`python3 scripts/test-release-workflows.py` exercises multi-commit pushes,
missing/shallow history, initial adoption, metadata failures, partial label
updates, retry and refusal to overwrite published assets. It uses temporary Git
repositories and a local fake `gh`; it creates no remote tags or releases.
`actionlint` 1.7.12 does not yet recognize GitHub's documented `concurrency.queue`
field; when using that version, ignore only its unsupported-key diagnostic for
`queue`, not other workflow errors. The setting is documented in
[GitHub workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#concurrency).
