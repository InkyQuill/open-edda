#!/usr/bin/env python3
"""Run a disposable production server for browser integration tests."""
import os
from pathlib import Path
import signal
import subprocess
import tempfile

root = Path(__file__).resolve().parent.parent
with tempfile.TemporaryDirectory(prefix="edda-browser-") as directory:
    binary = str(Path(directory) / "open-edda")
    subprocess.run(["go", "build", "-tags", "sqlite_fts5", "-o", binary, "."], cwd=root, check=True)
    env = dict(os.environ, OPEN_EDDA_ADDR="127.0.0.1:4187",
               OPEN_EDDA_DB_PATH=str(Path(directory) / "edda.db"),
               OPEN_EDDA_DATA_DIR=directory, OPEN_EDDA_MIGRATIONS_PATH=str(root / "migrations"),
               OPEN_EDDA_STATIC_PATH=str(root / "frontend/dist"),
               OPEN_EDDA_JWT_SECRET="browser-test-secret-at-least-32-bytes",
               OPEN_EDDA_API_KEY_ENCRYPTION_SECRET="browser-test-encryption-at-least-32-bytes",
               OPEN_EDDA_BOOTSTRAP_EMAIL="browser@example.invalid", OPEN_EDDA_BOOTSTRAP_PASSWORD="browser-test-password")
    server = subprocess.Popen([binary], cwd=root, env=env)
    def stop(_signal, _frame):
        server.terminate()
    signal.signal(signal.SIGTERM, stop)
    signal.signal(signal.SIGINT, stop)
    try:
        server.wait()
    finally:
        if server.poll() is None:
            server.terminate()
            server.wait(timeout=10)
