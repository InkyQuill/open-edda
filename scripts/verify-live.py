#!/usr/bin/env python3
"""Validate the authorized live deployment using a clearly marked test project.
Leaves that project for the author to inspect; temporary local checkouts are removed.
Never reads/uploads an existing author book. Credentials are read from a private file.
"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile
import time
import urllib.request

credentials = Path.home() / ".config/open-edda/bootstrap-login.json"
account = json.loads(credentials.read_text())
server = os.environ.get("EDDA_VERIFY_SERVER", account["server"])
cli_path = str(Path.home() / ".local/bin/edda")
env = {key: value for key, value in os.environ.items() if not key.startswith(("OPEN_EDDA_", "WRITER_"))}

def cli(*args, stdin=None, failure=False):
    result = subprocess.run([cli_path, *map(str, args)], env=env, input=stdin, text=True, capture_output=True)
    if (result.returncode != 0) != failure:
        raise RuntimeError(result.stderr or result.stdout)
    return result.stdout

with urllib.request.urlopen(server + "/api/health", timeout=20) as response:
    assert response.status == 200
cli("login", "--server", server, "--email", account["email"], "--password-stdin", stdin=account["password"] + "\n")
connection = json.loads((Path(os.environ.get("XDG_CONFIG_HOME", str(Path.home() / ".config"))) / "open-edda/client.json").read_text())
def api(method, path, payload=None):
    request = urllib.request.Request(server + path, data=None if payload is None else json.dumps(payload).encode(), method=method,
        headers={"Authorization": "Bearer " + connection["token"], "Content-Type": "application/json"})
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response)

title = os.environ.get("EDDA_VERIFY_TITLE", "Проверка синхронизации · 2026-10-06")
project = json.loads(cli("create", "--json", "--title", title))
project_id = project["id"]
base = f"/api/projects/{project_id}/files"
with tempfile.TemporaryDirectory(prefix="edda-live-check-") as directory:
    root = Path(directory)
    a, b = root / "first", root / "second"
    a.mkdir()
    (a / "chapter.md").write_text("# Проверочная глава\n\nЭта книга создана для проверки сервиса.\n")
    (a / "image.bin").write_bytes(bytes([0, 255, 13, 10, 1]))
    (a / "notes").mkdir()
    (a / "notes/.env").write_text("LOCAL_ONLY=1\n")
    cli("attach", a, "--project", project_id)
    cli("send", a)
    initial = api("GET", base + "/versions/current")
    cli("get", b, "--project", project_id)
    assert (a / "image.bin").read_bytes() == (b / "image.bin").read_bytes()
    assert not (b / "notes/.env").exists()
    cli("move", b, "--from", "chapter.md", "--to", "глава.md")
    cli("send", b)
    cli("take", a)
    assert (a / "глава.md").read_bytes() == (b / "глава.md").read_bytes()
    renamed = api("GET", base + "/versions/current")
    assert next(e["id"] for e in initial["entries"] if e["path"] == "chapter.md") == next(e["id"] for e in renamed["entries"] if e["path"] == "глава.md")
    (a / "глава.md").write_text("Локальный вариант — сохранить.\n")
    (b / "глава.md").write_text("Вариант второго компьютера.\n")
    cli("send", b)
    cli("take", a, failure=True)
    cli("resolve", a, "--path", "глава.md", "--use", "local")
    cli("take", a)
    cli("send", a)
    cli("take", b)
    assert (b / "глава.md").read_text() == "Локальный вариант — сохранить.\n"
    cli("restore", a, "--version", initial["id"])
    cli("take", a)
    restored = api("GET", base + "/versions/current")
    assert restored["entries"] == initial["entries"]
    subprocess.run(["ssh", "-p", "7777", "-o", "BatchMode=yes", "inky@direct.inkyquill.net", "kubectl -n open-edda rollout restart deployment/open-edda && kubectl -n open-edda rollout status deployment/open-edda --timeout=120s"], check=True)
    # Pod readiness can precede service endpoint propagation. Confirm the same
    # client route responds before starting a two-minute file-transfer request.
    for attempt in range(20):
        try:
            with urllib.request.urlopen(server + "/api/health", timeout=3) as response:
                if response.status == 200:
                    break
        except OSError:
            time.sleep(1)
    else:
        raise RuntimeError("Client route did not recover after pod restart")
    cli("take", b)
    assert (b / "chapter.md").read_bytes() == (a / "chapter.md").read_bytes()
    assert (b / "image.bin").read_bytes() == bytes([0, 255, 13, 10, 1])
    report = {"server": server, "title": title, "project": project_id, "initialVersion": initial["id"], "restoredVersion": restored["id"], "historyCount": len(json.loads(cli("history", a, "--json"))["versions"]), "binarySHA256": hashlib.sha256((b / "image.bin").read_bytes()).hexdigest(), "result": "PASS: authenticated sync, two copies, stable rename, conflicts, restore and real pod restart"}
    print(json.dumps(report, ensure_ascii=False, indent=2))
    if os.environ.get("EDDA_VERIFY_REPORT"):
        Path(os.environ["EDDA_VERIFY_REPORT"]).write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
