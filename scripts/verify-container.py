#!/usr/bin/env python3
"""Exercise a local image with isolated volumes; remove only this run's resources."""
import json
import os
from pathlib import Path
import socket
import subprocess
import sys
import tempfile
import time
import urllib.request
import uuid

image = sys.argv[1] if len(sys.argv) > 1 else "open-edda:sync-test"
repo = Path(__file__).resolve().parent.parent
prefix = "edda-verify-" + uuid.uuid4().hex[:10]
volumes = [prefix + "-source", prefix + "-restored"]
containers = []

def docker(*args):
    return subprocess.check_output(["docker", *args], text=True).strip()

def start(name, volume, restored=False, limited=False):
    env = ["OPEN_EDDA_JWT_SECRET=isolated-test-jwt-secret-at-least-32-bytes",
           "OPEN_EDDA_API_KEY_ENCRYPTION_SECRET=isolated-test-encryption-secret-at-least-32-bytes",
           "OPEN_EDDA_BOOTSTRAP_EMAIL=container@example.invalid",
           "OPEN_EDDA_BOOTSTRAP_PASSWORD=isolated-test-password"]
    if restored:
        env += ["OPEN_EDDA_DATA_DIR=/data/restored", "OPEN_EDDA_DB_PATH=/data/restored/edda.db"]
    with socket.socket() as listener:
        listener.bind(("127.0.0.1", 0))
        host_port = listener.getsockname()[1]
    args = ["run", "-d", "--name", name, "-p", f"127.0.0.1:{host_port}:8080"]
    args += ["--tmpfs", "/data:rw,size=8m,uid=10001,gid=10001"] if limited else ["-v", volume + ":/data"]
    for value in env:
        args += ["-e", value]
    docker(*args, image)
    containers.append(name)
    port = docker("port", name, "8080/tcp").rsplit(":", 1)[1]
    url = "http://127.0.0.1:" + port
    ready(url)
    return url

def ready(url):
    for _ in range(100):
        try:
            with urllib.request.urlopen(url + "/api/health", timeout=1) as response:
                if response.status == 200:
                    return
        except OSError:
            time.sleep(.1)
    raise RuntimeError("container did not become ready")

try:
    for volume in volumes:
        docker("volume", "create", volume)
    with tempfile.TemporaryDirectory(prefix=prefix) as directory:
        root = Path(directory)
        binary = root / "edda"
        subprocess.run(["go", "build", "-tags", "sqlite_fts5", "-o", str(binary), "./cmd/edda"], cwd=repo, check=True)
        env = {k: v for k, v in os.environ.items() if not k.startswith(("OPEN_EDDA_", "WRITER_"))}
        env["XDG_CONFIG_HOME"] = str(root / "config")
        def cli(*args, stdin=None):
            result = subprocess.run([str(binary), *map(str, args)], env=env, input=stdin, capture_output=True, text=True)
            if result.returncode:
                raise RuntimeError(result.stderr)
            return result.stdout
        url = start(prefix, volumes[0])
        cli("login", "--server", url, "--email", "container@example.invalid", "--password-stdin", stdin="isolated-test-password\n")
        project = json.loads(cli("create", "--json", "--title", "Container book"))
        a, b = root / "a", root / "b"
        a.mkdir()
        (a / "chapter.md").write_text("# Книга\n\n最初の章。\n")
        (a / "cover.bin").write_bytes(bytes([0, 255, 13, 10]))
        cli("attach", a, "--project", project["id"])
        cli("send", a)
        original = json.loads(cli("history", a, "--json"))["versions"][0]["id"]
        cli("get", b, "--project", project["id"])
        cli("move", b, "--from", "chapter.md", "--to", "renamed.md")
        (b / "renamed.md").write_text("Edited in a second copy.\n")
        cli("send", b)
        docker("kill", prefix)
        docker("start", prefix)
        ready(url)
        cli("take", a)
        assert (a / "renamed.md").read_text() == "Edited in a second copy.\n"
        assert not (a / "chapter.md").exists()
        docker("exec", prefix, "mkdir", "-p", "/data/backups")
        docker("exec", prefix, "edda", "backup", "--db", "/data/edda.db", "--data", "/data", "--output", "/data/backups/snapshot")
        docker("exec", prefix, "edda", "verify-backup", "--source", "/data/backups/snapshot")
        docker("run", "--rm", "-v", volumes[0] + ":/source:ro", "-v", volumes[1] + ":/data", "--entrypoint", "edda", image,
               "restore-backup", "--source", "/source/backups/snapshot", "--output", "/data/restored")
        restored_url = start(prefix + "-restored", volumes[1], True)
        cli("login", "--server", restored_url, "--email", "container@example.invalid", "--password-stdin", stdin="isolated-test-password\n")
        c = root / "restored"
        cli("get", c, "--project", project["id"])
        assert (c / "renamed.md").read_bytes() == (a / "renamed.md").read_bytes()
        assert (c / "cover.bin").read_bytes() == bytes([0, 255, 13, 10])
        cli("restore", c, "--version", original)
        cli("take", c)
        assert (c / "chapter.md").read_text() == "# Книга\n\n最初の章。\n"
        assert not (c / "renamed.md").exists()
        evidence = {
            "sourceBase": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=repo, text=True).strip(),
            "sourceState": "working tree including current uncommitted file/sync changes",
            "image": docker("image", "inspect", image, "--format", "{{.Id}}"),
            "project": project["id"], "originalVersion": original,
            "restoredHistory": [v["id"] for v in json.loads(cli("history", c, "--json"))["versions"]],
        }
        # A real ENOSPC during upload must keep the acknowledged head and the
        # pending client snapshot; freeing space makes the same send retryable.
        full_url = start(prefix + "-full", "", limited=True)
        cli("login", "--server", full_url, "--email", "container@example.invalid", "--password-stdin", stdin="isolated-test-password\n")
        full_project = json.loads(cli("create", "--json", "--title", "Full volume"))
        full_copy = root / "full"
        cli("get", full_copy, "--project", full_project["id"])
        old_state = json.loads((full_copy / ".edda/checkout.json").read_text())
        available = int(docker("exec", prefix + "-full", "df", "-B1", "--output=avail", "/data").splitlines()[-1])
        docker("exec", prefix + "-full", "dd", "if=/dev/zero", "of=/data/filler", "bs=1024", "count=" + str((available - 1024 * 1024) // 1024))
        (full_copy / "large.bin").write_bytes(b"x" * (2 * 1024 * 1024))
        failed = subprocess.run([str(binary), "send", str(full_copy)], env=env, text=True, capture_output=True)
        assert failed.returncode != 0
        failed_state = json.loads((full_copy / ".edda/checkout.json").read_text())
        assert failed_state["base"]["id"] == old_state["base"]["id"] and failed_state.get("pending")
        assert json.loads(cli("history", full_copy, "--json"))["versions"][0]["id"] == old_state["base"]["id"]
        docker("exec", prefix + "-full", "rm", "/data/filler")
        cli("send", full_copy)
        verified_copy = root / "full-verified"
        cli("get", verified_copy, "--project", full_project["id"])
        assert (verified_copy / "large.bin").read_bytes() == (full_copy / "large.bin").read_bytes()
        evidence["diskFullRecovery"] = "upload rejected without head change; pending retry succeeded after freeing space"
        cli("logout")
        print(json.dumps(evidence, indent=2))
        print("PASS: non-root container, persistent volume, two-copy sync/rename, abrupt restart, online backup, fresh-volume restore, full version history, exact bytes and ENOSPC retry")
except Exception:
    for name in containers:
        subprocess.run(["docker", "logs", "--tail", "15", name], check=False)
    raise
finally:
    for name in reversed(containers):
        subprocess.run(["docker", "rm", "-f", name], stdout=subprocess.DEVNULL, check=False)
    for volume in reversed(volumes):
        subprocess.run(["docker", "volume", "rm", volume], stdout=subprocess.DEVNULL, check=False)
