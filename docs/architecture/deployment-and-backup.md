# Container deployment and verified backups

Implemented and exercised locally on 2026-10-06. The Dockerfile builds the frontend and both Go executables from source, uses the frozen Bun lockfile, pins base-image digests and runs as UID/GID 10001. The service is a single instance with SQLite WAL and immutable objects. Package repository updates can still change OS package bytes; this is not a claim of bit-for-bit image reproducibility.

## Start

Supply `OPEN_EDDA_JWT_SECRET`, `OPEN_EDDA_API_KEY_ENCRYPTION_SECRET`, `OPEN_EDDA_BOOTSTRAP_EMAIL` and `OPEN_EDDA_BOOTSTRAP_PASSWORD` through the environment or a private Compose environment file. Generate independent random secrets. Keep the encryption secret for restoring encrypted provider settings; it is not included in a data backup. Bootstrap creates the initial account only; it is not a password-reset mechanism.

```sh
docker compose up --build -d
```

The example binds localhost port 8080. Put an HTTPS reverse proxy in front for remote access. The named volume mounts at `/data`; authoritative data is `/data/edda.db` (and SQLite WAL/SHM while running) plus `/data/objects`. Frontend and migrations are image files. For bind mounts, grant UID/GID 10001 access before starting. Do not edit immutable objects directly.

`deploy/kubernetes.yaml` provides a single-replica Recreate deployment and a ReadWriteOncePod claim. Replace the image with a published registry digest, choose a suitable storage class and create `open-edda-secrets` with the same environment keys. A block-backed filesystem with SQLite-compatible locking/fsync semantics is required. No shared NFS/SMB database or multiple active replicas are supported. The manifest is a deployment template; no real Kubernetes cluster/PVC has been qualified in this task.

## Online backup

The `edda` executable is included in the image. The output must be a new directory under an existing parent; keep backups on a separate protected volume or copy completed backups off-host.

```sh
edda backup --db /data/edda.db --data /data --output /backups/book-2026-10-06
edda verify-backup --source /backups/book-2026-10-06
```

Backup creates a consistent SQLite snapshot with `VACUUM INTO`, then copies all objects referenced by all versions in that snapshot. Immutable objects are never garbage-collected in this release, so the source server can keep writing during the copy. Backup does not copy a live main database file. It verifies SQLite integrity and foreign keys, object sizes and SHA-256, and records the database checksum and object inventory in `backup.json`. Partial output stays in a temporary sibling and is never installed as a completed backup. Failure leaves the source untouched.

Backups contain author content, account records and encrypted provider settings; protect them as private data. External encryption/JWT secrets are operational configuration and must be backed up separately. A checksum is corruption detection, not a signature against malicious modification.

## Restore

```sh
edda restore-backup --source /backups/book-2026-10-06 --output /data/restored
```

Restore verifies the backup and copies it into a new root, refusing any existing destination. It does not overwrite a running instance. Start a separate service with `OPEN_EDDA_DATA_DIR=/data/restored` and `OPEN_EDDA_DB_PATH=/data/restored/edda.db`, supply the original encryption secret, then verify projects/history/downloads before switching traffic. Keep the previous deployment/data root available until the restored instance is accepted. Server credentials are the account records from the snapshot.

There is no automatic version/object/recovery-snapshot deletion. Retention and deletion are explicit administration decisions. `edda take` recovery snapshots live inside each local checkout and are separate from server backups.

## Reproducible acceptance check

```sh
docker build -t open-edda:sync-test .
python3 scripts/verify-container.py open-edda:sync-test
```

The check creates uniquely named disposable containers/volumes, builds the real CLI, creates a project, attaches a folder, sends/downloads Unicode text and binary data, renames in a second copy, kills/restarts the server, makes an online backup, restores into a separate fresh volume, and retrieves an older version. It verifies exact contents and history, exercises actual ENOSPC on a limited tmpfs followed by a successful retry, then removes only its own containers, volumes and temporary credentials. It passed on the local Docker filesystem; hardware power failure and a production CSI driver require environment-specific qualification.
