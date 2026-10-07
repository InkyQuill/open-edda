#!/usr/bin/env python3
"""Exercise release scripts with local Git history and a fault-injected gh stub."""
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parent.parent

GH_STUB = '''#!/usr/bin/env python3
import json, os, sys
from pathlib import Path
path=Path(os.environ['FAKE_STATE'])
state=json.loads(path.read_text())
args=sys.argv[1:]
command=' '.join(args)
state['calls'].append(command)
fail=os.environ.get('FAIL_AT','')
if fail and fail in command:
 path.write_text(json.dumps(state)); sys.exit(1)
if args[:2]==['release','view']: print('false' if state['published'] else 'true')
elif args[:2]==['release','edit']: state['published']=True
elif args[0]=='api' and '/pulls' in args[1]: print(42)
elif args[0]=='api' and '/labels' in args[1]: print(1 if state['pending'] else 0)
elif args[:2]==['pr','edit'] and '--remove-label' in args: state['pending']=False
elif args[:2]==['pr','edit'] and '--add-label' in args: state['tagged']=True
path.write_text(json.dumps(state))
'''


class ReleaseScripts(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix='edda-release-test-')
        self.addCleanup(self.temp.cleanup)
        self.work = Path(self.temp.name)

    def run_script(self, name, env=None, cwd=None):
        return subprocess.run(['bash', str(ROOT / 'scripts' / name)],
                              cwd=cwd or self.work, env=dict(os.environ, **(env or {})),
                              capture_output=True, text=True)

    def test_failed_label_steps_leave_draft_and_allow_retry(self):
        for fail in ['label create', 'commits/v0.1.0/pulls', '--add-label', '--remove-label', 'release edit']:
            with self.subTest(fail=fail):
                self.check_publish_retry(fail)

    def check_publish_retry(self, fail):
        stub = self.work / 'gh'
        stub.write_text(GH_STUB.replace('#!/usr/bin/env python3', '#!' + sys.executable, 1))
        stub.chmod(0o700)
        state_file = self.work / 'state.json'
        state_file.write_text(json.dumps(dict(published=False, pending=True, tagged=False, calls=[])))
        for arch in ['amd64', 'arm64']:
            (self.work / f'edda-v0.1.0-linux-{arch}.tar.gz').write_bytes(arch.encode())
        (self.work / 'unrelated.tar.gz').write_bytes(b'not a release asset')
        env = dict(PATH=str(self.work)+os.pathsep+os.environ['PATH'], TAG='v0.1.0',
                   GH_REPO='example/edda', FAKE_STATE=str(state_file), FAIL_AT=fail)
        first = self.run_script('publish-cli-release.sh', env)
        self.assertNotEqual(first.returncode, 0)
        self.assertFalse(json.loads(state_file.read_text())['published'])
        env['FAIL_AT'] = ''
        retry = self.run_script('publish-cli-release.sh', env)
        self.assertEqual(retry.returncode, 0, retry.stderr)
        state = json.loads(state_file.read_text())
        self.assertTrue(state['published'])
        self.assertNotIn('unrelated.tar.gz', (self.work / 'SHA256SUMS').read_text())
        self.assertFalse(any('unrelated.tar.gz' in call for call in state['calls']))
        self.assertTrue(state['tagged'])
        self.assertFalse(state['pending'])
        self.assertIn('release edit', state['calls'][-1])
        previous_uploads = sum('release upload' in call for call in state['calls'])
        refused = self.run_script('publish-cli-release.sh', env)
        self.assertNotEqual(refused.returncode, 0)
        state = json.loads(state_file.read_text())
        self.assertEqual(sum('release upload' in call for call in state['calls']), previous_uploads)

    def git(self, cwd, *args):
        return subprocess.check_output(['git', '-C', str(cwd), *args], stderr=subprocess.DEVNULL, text=True).strip()

    def commit(self, repo, name, data):
        (repo / name).write_text(data)
        self.git(repo, 'add', name)
        self.git(repo, '-c', 'user.name=Release Test', '-c', 'user.email=release@example.invalid',
                 'commit', '-m', 'test: change '+name)
        return self.git(repo, 'rev-parse', 'HEAD')

    def test_multi_commit_push_requires_history_and_detects_version_change(self):
        repo = self.work / 'origin'
        repo.mkdir()
        self.git(repo, 'init', '-b', 'main')
        initial = self.commit(repo, 'README.md', 'start')
        before = self.commit(repo, 'version.txt', '0.0.0\n')
        self.commit(repo, 'version.txt', '0.1.0\n')
        same_version = self.commit(repo, 'README.md', 'middle')
        self.commit(repo, 'README.md', 'last')
        clone = self.work / 'clone'
        self.git(self.work, 'clone', '--depth', '2', repo.as_uri(), str(clone))
        env = dict(GITHUB_EVENT_NAME='push', BEFORE=before)
        shallow = self.run_script('release-version-changed.sh', env, clone)
        self.assertNotEqual(shallow.returncode, 0, 'Missing history must not look like no change')
        self.git(clone, 'fetch', '--unshallow')
        full = self.run_script('release-version-changed.sh', env, clone)
        self.assertEqual((full.returncode, full.stdout.strip()), (0, 'true'))
        env['BEFORE'] = same_version
        self.assertEqual(self.run_script('release-version-changed.sh', env, clone).stdout.strip(), 'false')
        # Initial adoption has no version file on the previous commit.
        env['BEFORE'] = initial
        self.git(clone, 'checkout', before)
        adopted = self.run_script('release-version-changed.sh', env, clone)
        self.assertEqual((adopted.returncode, adopted.stdout.strip()), (0, 'false'))
        env['BEFORE'] = '0'*40
        self.assertEqual(self.run_script('release-version-changed.sh', env, clone).stdout.strip(), 'false')
        env.update(GITHUB_EVENT_NAME='workflow_dispatch', BEFORE='')
        self.assertEqual(self.run_script('release-version-changed.sh', env, clone).stdout.strip(), 'false')


if __name__ == '__main__':
    unittest.main()
