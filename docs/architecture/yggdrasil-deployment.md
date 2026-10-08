# Yggdrasil deployment

Deployed to k3s on 2026-10-06 from the working tree; external nginx host activation requires the one-time sudo step below. Host: `ssh -p 7777 inky@direct.inkyquill.net`.

- Public origin: `https://edda.inky.su`.
- Namespace/deployment/service: `open-edda`.
- Image: `docker.io/library/open-edda:20261006-files-sync`, imported into k3s; `imagePullPolicy: Never`.
- Persistent data: `open-edda-data`, 10 GiB `local-path` RWO PVC, single pod on yggdrasil; Recreate rollout. Host storage is local Btrfs on NVMe. The request is not an enforced per-directory quota in local-path.
- Existing nginx terminates TLS using `/etc/nginx/ssl.inky.su` (wildcard certificate). A new virtual host forwards to the existing Istio gateway on NodePort 30880. No other host is changed.
- HTTPRoute: `edda.inky.su` → `open-edda:8080`, through `istio-ingress/ingress-gw`.
- Kubernetes Secret `open-edda-secrets` contains bootstrap and application secrets. Values are not in the repository.

Configuration is in `deploy/yggdrasil.yaml` and `deploy/edda.inky.su.nginx.conf`. The one-time `activate-yggdrasil-host.sh` verifies staged SHA-256 checksums, imports the image and installs the new nginx host only if no Edda host already exists. It checks nginx before reload and removes its new config if validation fails. The host requires interactive sudo for this step; only nginx reload is passwordless. The image has already been imported using the account’s existing containerd group access; repeating that import in the activation script is harmless.

The local CLI is installed at `~/.local/bin/edda`. Initial web credentials are saved privately in `~/.config/open-edda/bootstrap-login.json` (0600). Once login succeeds, the CLI stores its normal scoped token in `~/.config/open-edda/client.json`; bootstrap secrets do not enter a book directory.

## Live verification

`scripts/verify-live.py` uses the installed CLI and private bootstrap credentials to create a clearly marked test project. It tests HTTPS, attachment, exact Unicode/binary bytes, exclusions, identity-preserving rename, concurrent edits, resolution, historical restore and a Kubernetes rollout restart. It leaves the remote test project for inspection and removes temporary local checkouts. Run only when a live test project and a brief single-pod restart are intended.

## Frontend update — 2026-10-07

The public host is active. Previous image: `docker.io/library/open-edda:20261007-quiet-ui-v1` (index `sha256:bcd34587a98fafeacb5beb000524eb58f1b0f9859c3c1c6509dfa9de73f2bf52`). This image adds only frontend assets to the previous deployment image; server, CLI and migrations remain unchanged. The initial-deployment notes below are historical. See [UI acceptance and rollback](../design/production-acceptance.md).

## Ignore rules and project deletion — 2026-10-07

Initial fix image: `docker.io/library/open-edda:20261007-ignore-delete`, built from the working tree (server, CLI and frontend). Image ID: `sha256:9c2ffd7f721e6dec3fbab99d2430a773be1467b619a82dea372dc17199ffaf6f`. Transferred archive SHA-256: `9db6f6e4a39ae0d3a625536d9880c7b74b0ba3bde474c94ba26057aa4d557e6c`; verified before import. No schema change.

Before rollout, a consistent backup was made and verified at `/data/backups/pre-ignore-delete-20261007-1837`, then exported privately to `~/.local/share/open-edda/backups/pre-ignore-delete-20261007-1837.tar.gz` (SHA-256 `553d7c698b50aac7668e636e8197aa03d6116e23dc2aabb8d9169ebab3f012af`). Rollout and public HTTPS health passed. A disposable project was created and deleted through the live confirmation dialog; all four pre-existing project version heads remained unchanged. The installed local CLI was updated and its previous binary retained in `~/.local/share/open-edda/backups/edda-before-ignore-20261007`.

Follow-up current image: `docker.io/library/open-edda:20261007-ignore-rm` adds automatic external-symlink exclusion and `edda rm` / `edda rm --undo`. Image index: `sha256:b6b97f7f77bbdf9874970795dae7b04bbb007468a3ffe51a8e9d8340eb8f1cd8`; transferred archive SHA-256: `4bae3b5f1db7105f1d55f6e31008a6608fbe2b9df7fe69ec91785483ca1b0774`. Full Go race tests/vet passed again; the installed CLI and container expose the new command.

Rollback image: `docker.io/library/open-edda:20261007-quiet-ui-v1`. Keep the PVC and backups; changing the image does not undo a confirmed project deletion.

## Maintenance

For an update, build a unique image tag, save/transfer it to the host, import it into k3s with sudo, update the image in the manifest and apply it. Keep the previous imported image/tag available for rollback. Do not reuse the one-time nginx activation script: it intentionally refuses existing host configuration.

Before an update, make and verify a consistent online backup using the container's `edda` command; export completed backups off the data volume. Keep the encryption secret separately. PVC deletion is data destruction; rolling back a Deployment must never delete the PVC. See [backup/restore](deployment-and-backup.md).

## Initial deployment evidence

The PVC is bound (`pvc-ff7b2d11-7b70-421c-8928-3ea6be9cfb88`); the deployment runs as UID/GID 10001 and the existing Istio HTTP route answers `/api/health` with status `ok`. Image index: `sha256:996a3fde8c658eb4c6070b8fb9d340803e5cf9d8210469791963025ddf4157b9`. Installed CLI SHA-256: `43911409677857ced5eab1fc61c365cb426576e80a93842a6fdd2078764e7a1e`.

An initial consistent backup was verified in the pod at `/data/backups/initial-20261006` and exported privately to `~/.local/share/open-edda/backups/initial-20261006.tar.gz` on the workstation. Application secrets were preserved separately in `~/.config/open-edda/yggdrasil-secrets.json` (0600).

During TLS preflight, the active wildcard certificate was found to expire on October 6, despite an already renewed certificate in the account’s acme.sh directory. Certificate/key public hashes were checked for equality; the previous deployed pair was backed up privately on the host under `.local/share/open-edda-deploy/20261006/tls-backup`. The renewed pair was installed through acme.sh with the existing passwordless nginx reload command, also recording that reload for subsequent acme.sh installs. Public TLS on `vault.inky.su` then verified the renewed expiry of December 5, 2026. No fresh certificate issuance or DNS change was needed.

Until `/etc/nginx/sites-enabled/edda.inky.su` is installed, SNI for Edda reaches the host’s default virtual host and does not pass hostname verification. Do not bypass certificate validation in the CLI. Activate with:

```sh
ssh -t -p 7777 inky@direct.inkyquill.net 'sudo /home/inky/.local/share/open-edda-deploy/20261006/activate-yggdrasil-host.sh'
```

The prepared script and nginx config are in this repository and on the host; SHA-256 checks passed on the transferred image and both files. Public HTTPS/browser acceptance must run after activation.

The installed CLI completed the real-cluster round-trip over an encrypted SSH tunnel: project `project-1791291816394618365-1`, initial version `355c1bdccca944a2a67b54fe6b3524dd`, restored version `462c17486601c98fa12d87121c86c778`, six versions retained. Binary SHA-256 after restart: `24fcff736fd1e6b0e1d829f71f81d8703b2904df3f0dafada69c297ffe527dfe`. The first immediate post-rollout request hit an endpoint-propagation timeout; the acceptance script now waits for the client route after pod readiness. The second full run passed without changing application code. Two explicitly named test projects remain for inspection; no real author folder was uploaded.

A post-test backup at `/data/backups/verified-20261006` was verified, exported to `~/.local/share/open-edda/backups/verified-20261006.tar.gz`, and restored independently on the workstation. SQLite and all referenced object checks passed; 12 project versions were retained across the two test projects. The temporary restored copy was removed, leaving the verified archive. These are actual local-path PVC/restart/backup checks, not a hardware-power-loss qualification.

## Refresh sessions — 2026-10-07

Current image: `docker.io/library/open-edda:20261007-sessions-secure`, locally
built and imported; no registry upload. OCI index
`sha256:ad1cecdf50e514a5a334b3802ab107bfd1dcd11ca3680b641f216c78ea6bfc2e`;
archive SHA-256 `d8e2f9821be7d8ccb8dc3985e171c3cff85e6b29a92179f06a53efbc8a5c74b4`.
Schema 9 adds hashed refresh sessions. See [session behavior](sessions.md).

Pre-migration online backup `/data/backups/pre-sessions-20261007-1920` was verified
and exported privately to the workstation (archive SHA-256
`f19e1f3bbf620cb7caec046479b82b52c4cd6766b6d3177a37b1fd785f975b24`).
The previous image `20261007-ignore-rm` is retained. Rolling back to it stops
refresh support; it does not require dropping the new table or restoring old data.

Live HTTPS acceptance passed: Secure/HttpOnly/SameSite refresh cookie, silent
access recovery, server-side refresh revocation on logout, invalid-session login
redirect and distinct appearance/service icons. These checks changed no projects.
