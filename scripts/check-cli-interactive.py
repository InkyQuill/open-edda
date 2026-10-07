#!/usr/bin/env python3
"""Linux PTY acceptance check against an isolated real Edda server.

Usage: python3 scripts/check-cli-interactive.py CLI_BINARY SERVER_BINARY
Only temporary folders, credentials and a localhost server are used.
"""
import errno
import json
import os
from pathlib import Path
import pty
import select
import signal
import shutil
import socket
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request

CLI, SERVER = map(lambda p: str(Path(p).resolve()), sys.argv[1:])
REPO = Path(__file__).resolve().parents[1]


def command(env, *args, cwd=None, ok=True):
    result = subprocess.run([CLI, *map(str, args)], env=env, cwd=cwd,
                            stdin=subprocess.DEVNULL, capture_output=True,
                            text=True, timeout=15)
    if ok:
        assert result.returncode == 0, (args, result.stdout, result.stderr)
    else:
        assert result.returncode != 0, args
    return result.stdout


def dialogue(env, args, steps, cwd=None, ok=True):
    pid, fd = pty.fork()
    if pid == 0:
        if cwd:
            os.chdir(cwd)
        os.execve(CLI, [CLI, *map(str, args)], env)
    transcript = b""
    pending = b""
    deadline = time.monotonic() + 15
    status = None

    def read():
        nonlocal transcript, pending
        if time.monotonic() > deadline:
            raise AssertionError((args, "terminal timeout", transcript.decode(errors="replace")))
        if select.select([fd], [], [], .1)[0]:
            try:
                data = os.read(fd, 65536)
            except OSError as exc:
                if exc.errno != errno.EIO:
                    raise
                return False
            if not data:
                return False
            transcript += data
            pending += data
        return True

    try:
        for prompt, answer in steps:
            needle = prompt.encode()
            while needle not in pending:
                assert read(), (args, "missing prompt", prompt, transcript)
            pending = pending.split(needle, 1)[1]
            os.write(fd, answer.encode())
        while read():
            pass
        _, status = os.waitpid(pid, 0)
        assert (os.waitstatus_to_exitcode(status) == 0) == ok, (args, transcript)
        return transcript.decode(errors="replace")
    finally:
        os.close(fd)
        if status is None:
            try:
                os.kill(pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            os.waitpid(pid, 0)


with tempfile.TemporaryDirectory(prefix="edda-cli-acceptance-") as directory:
    root = Path(directory)
    static = root / "static"
    static.mkdir()
    (static / "index.html").write_text("Edda test")
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        port = sock.getsockname()[1]
    url = f"http://127.0.0.1:{port}"
    env = {k: v for k, v in os.environ.items()
           if not k.startswith(("OPEN_EDDA_", "WRITER_"))}
    env.update(XDG_CONFIG_HOME=str(root / "config"),
               OPEN_EDDA_ADDR=f"127.0.0.1:{port}",
               OPEN_EDDA_SECRET="isolated-acceptance-secret-at-least-32-bytes",
               OPEN_EDDA_API_KEY_ENCRYPTION_SECRET="isolated-encryption-secret-at-least-32-bytes",
               OPEN_EDDA_DB_PATH=str(root / "data" / "edda.db"),
               OPEN_EDDA_DATA_DIR=str(root / "data"),
               OPEN_EDDA_MIGRATIONS_PATH=str(REPO / "migrations"),
               OPEN_EDDA_STATIC_PATH=str(static),
               OPEN_EDDA_BOOTSTRAP_EMAIL="test@example.invalid",
               OPEN_EDDA_BOOTSTRAP_PASSWORD="isolated-test-password")
    (root / "data").mkdir()
    with (root / "server.log").open("w") as log:
        server = subprocess.Popen([SERVER], env=env, stdout=log, stderr=log)
        try:
            for _ in range(100):
                try:
                    urllib.request.urlopen(url, timeout=.2).close()
                    break
                except (OSError, urllib.error.URLError):
                    assert server.poll() is None, (root / "server.log").read_text()
                    time.sleep(.05)
            else:
                raise AssertionError("test server did not start")

            output = dialogue(env, ["login"], [
                ("Server URL:", url + "\n"),
                ("Email:", "test@example.invalid\n"),
                ("Password:", "isolated-test-password\n")])
            assert "isolated-test-password" not in output
            source = root / "book"
            source.mkdir()
            (source / "chapter.md").write_text("Original chapter\n")
            (source / "local-only.txt").write_text("Keep local\n")
            dialogue(env, ["send", "--exclude", "local-only.txt"], [
                ("Project folder", "\n"),
                ("(new / existing)", "new\n"),
                ("New project title", "Книга\n")], cwd=source)
            state = json.loads((source / ".edda" / "checkout.json").read_text())
            original = state["base"]["id"]
            dialogue(env, ["send", source], [])
            assert not any(x["path"] == "local-only.txt" for x in state["base"]["entries"])
            assert "Книга" in command(env, "projects")
            projects = json.loads(command(env, "projects", "--json"))
            assert len(projects) == 1
            other = root / "other"
            dialogue(env, ["get"], [("New destination folder", str(other) + "\n"),
                                     ("Project number or ID", "1\n")])
            assert (other / "chapter.md").read_text() == "Original chapter\n"
            assert not (other / "local-only.txt").exists()
            command(env, "get", other, "--project", projects[0]["id"], ok=False)
            dialogue(env, ["status"], [], cwd=other)
            dialogue(env, ["move", other], [("Source relative path", "chapter.md\n"),
                                            ("Destination relative path", "renamed.md\n")])
            command(env, "send", other)
            dialogue(env, ["take"], [], cwd=source)
            assert (source / "renamed.md").exists()
            (source / "renamed.md").write_text("Local edit\n")
            (other / "renamed.md").write_text("Remote edit\n")
            command(env, "send", other)
            command(env, "take", source, ok=False)
            assert "Conflict" in command(env, "conflicts", source)
            dialogue(env, ["resolve", source], [("Conflict path", "renamed.md\n"),
                                                ("Use version", "remote\n")])
            command(env, "take", source)
            assert (source / "renamed.md").read_text() == "Remote edit\n"
            assert "VERSION" in command(env, "history", source)
            assert len(json.loads(command(env, "history", source, "--json"))["versions"]) >= 3
            dialogue(env, ["restore", source], [("Version ID", original + "\n")])
            command(env, "take", source)
            assert (source / "chapter.md").read_text() == "Original chapter\n"
            assert (source / "local-only.txt").read_text() == "Keep local\n"
            oneoff = root / "oneoff"
            oneoff.mkdir()
            (oneoff / "note.md").write_text("Import me\n")
            assert "1 files" in command(env, "import", oneoff, "--dry-run")
            assert json.loads(command(env, "import", oneoff, "--dry-run", "--json"))["entries"]
            dialogue(env, ["create"], [("Project title", "One-time\n")])
            projects = json.loads(command(env, "projects", "--json"))
            one_id = next(p["id"] for p in projects if p["title"] == "One-time")
            dialogue(env, ["import", oneoff], [("Project number or ID", "\x03")], ok=False)
            dialogue(env, ["import", oneoff], [("Project number or ID", one_id + "\n")])
            assert (oneoff / ".edda" / "checkout.json").exists()
            nested = oneoff / "nested" / "inside"
            nested.mkdir(parents=True)
            assert one_id in dialogue(env, ["status"], [], cwd=nested)
            dialogue(env, ["send"], [], cwd=nested)
            dialogue(env, ["take"], [], cwd=nested)
            assert not (nested / ".edda").exists()
            second = root / "second"
            second.mkdir()
            (second / "note.md").write_text("Attached\n")
            created = json.loads(command(env, "create", "--title", "Attached", "--json"))
            dialogue(env, ["attach", second], [("Project number or ID", created["id"] + "\n")])
            command(env, "send", second)
            # Cancellation before project creation must leave source and server unchanged.
            cancelled = root / "cancelled"
            cancelled.mkdir()
            before = command(env, "projects", "--json")
            dialogue(env, ["send", cancelled], [("(new / existing)", "\x04")], ok=False)
            dialogue(env, ["send", cancelled], [("(new / existing)", "new\n"),
                                                ("New project title", "\x03")], ok=False)
            assert not (cancelled / ".edda").exists()
            assert command(env, "projects", "--json") == before
            backup = root / "backup"
            dialogue(env, ["backup"], [("SQLite database file", env["OPEN_EDDA_DB_PATH"] + "\n"),
                                       ("Object data directory", env["OPEN_EDDA_DATA_DIR"] + "\n"),
                                       ("New backup directory", str(backup) + "\n")])
            dialogue(env, ["verify-backup"], [("Backup directory", str(backup) + "\n")])
            dialogue(env, ["restore-backup"], [("Backup directory", str(backup) + "\n"),
                                               ("New restored data directory", str(root / "restored") + "\n")])
            local = root / "prototype"
            shutil.copytree(REPO / "fileproject" / "testdata" / "partial", local)
            dialogue(env, ["init"], [("Project folder", "\n"),
                                      ("Project title", "Local prototype\n")], cwd=local)
            dialogue(env, ["ids"], [("IDs action", "sync\n"),
                                     ("Project folder", "\n")], cwd=local)
            dialogue(env, ["checkpoint"], [("Project folder", "\n"),
                                            ("Checkpoint note", "Original\n")], cwd=local)
            checkpoints = json.loads(command(env, "history", local, "--json"))
            checkpoint = checkpoints[0]["id"]
            original_local = (local / "story" / "chapter-01.md").read_text()
            (local / "story" / "chapter-01.md").write_text("Changed locally\n")
            dialogue(env, ["diff", local], [("Source checkpoint ID", checkpoint + "\n")])
            dialogue(env, ["save"], [("Project folder", "\n"),
                                      ("Checkpoint note", "Changed\n")], cwd=local)
            dialogue(env, ["files"], [("Project folder", "\n")], cwd=local)
            dialogue(env, ["restore", local], [("Checkpoint ID", checkpoint + "\n")])
            assert (local / "story" / "chapter-01.md").read_text() == original_local
            # Every command has useful help without input, connection or side effects.
            for name in ("login logout projects create attach get send take status history restore "
                         "conflicts resolve move import backup verify-backup restore-backup init "
                         "ids save checkpoint files diff").split():
                assert "edda " in command(env, name, "--help")
                assert command(env, "help", name) == command(env, name, "--help")
            command(env, "logout")
            print("PASS: real server + terminal login; first send; project selection; get; status; move; "
                  "send/take; conflicts/resolve; history/restore; create/attach/import; human/JSON output; "
                  "exclusion preservation; cancellation; backup/verify/restore; local prototype workflows; all command help; logout")
        finally:
            server.terminate()
            server.wait(timeout=5)
