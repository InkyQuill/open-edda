package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var commandHelp = map[string]string{
	"login": `Log in and save the connection for future commands.
  edda login [--server URL] [--email EMAIL]
  edda login --server URL --email EMAIL --password-stdin
Missing values are requested; the password is hidden. --password-stdin is for pipes.`,
	"logout": `Remove the locally saved login.
  edda logout
This does not revoke sessions on the server or change environment variables.`,
	"projects": `List project titles and IDs.
  edda projects [--json]
--json returns the full project list for scripts.`,
	"create": `Create an empty file project on the server. No local files are uploaded.
  edda create [--title TITLE] [--server URL] [--json]
The title is requested if omitted. To upload a local folder, use edda send FOLDER.`,
	"send": `Upload local files to Edda, then send subsequent changes from the same folder.
  edda send [FOLDER]
  edda send FOLDER --title "My book"
  edda send FOLDER --project ID
For a folder not yet attached, the dialogue offers a new or existing project.
--title creates a new project; --project selects an existing one. These first-send
options also accept --server URL and repeatable --exclude relative/path.
An existing nonempty project must match the local folder before attachment.
The folder is remembered through .edda metadata. Later, just run edda send FOLDER.
To preview files and exclusions without uploading: edda import FOLDER --dry-run.`,
	"attach": `Connect an existing local folder to a server project without uploading it.
  edda attach [FOLDER] [--project ID] [--server URL] [--exclude relative/path]
Select a project from the list if --project is omitted. Then use edda send FOLDER.
A nonempty remote must match the local files exactly. --exclude is repeatable.`,
	"get": `Download a server project into a NEW local folder.
  edda get [NEW_DIRECTORY] [--project ID] [--server URL] [--version ID]
Missing destination and project are requested. Existing folders are never replaced.
--version selects a historical version; the default is the current version.`,
	"take": `Bring server changes into a connected local folder.
  edda take [FOLDER] [--restart]
Local edits are merged; conflicting files require edda conflicts and edda resolve.
Run take again after resolving them. --restart rebuilds an unapplied update plan
when local files have changed; recovery snapshots are retained.`,
	"status": `Show local project state, pending operations and unsent changes.
  edda status [FOLDER]
This works offline and does not check for new server changes.`,
	"history": `List versions of a connected project or checkpoints of a local prototype.
  edda history [FOLDER] [--cursor ID] [--json]
  edda history [FOLDER] --id FILE_ID
--cursor requests older server versions. --id filters local checkpoint history.`,
	"restore": `Restore a saved version. Omitted version/checkpoint is requested.
  edda restore [FOLDER] [--version ID]
For connected folders, publishes the chosen version on the server; use edda take
next to reconcile local files. Run edda history to inspect available versions.
  edda restore [LOCAL_FOLDER] [--checkpoint ID]
For local prototypes, restores a checkpoint directly into local files.`,
	"conflicts": `Show unresolved conflicts and recovery information.
  edda conflicts [FOLDER]
Choose a version with edda resolve, then run edda take for connected folders.`,
	"resolve": `Choose how to resolve a conflict; omitted required choices are requested.
  edda resolve [FOLDER] [--path PATH] [--use local|remote]
Use . as the path for a whole-tree conflict. Then run edda take.
For local prototypes:
  edda resolve [FOLDER] [--id FILE_ID] [--use local|server | --body-file FILE]`,
	"move": `Move a file or directory locally while preserving its server identity.
  edda move [FOLDER] [--from relative/path] [--to relative/path]
Missing paths are requested. Destination parents must exist; existing files are
never replaced. Run edda send afterwards. Repeating move resumes an interrupted move.`,
	"import": `Upload files into an EMPTY server project and save its binding in .edda.
  edda import [FOLDER] [--project ID] [--server URL] [--exclude relative/path] [--json]
  edda import [FOLDER] --dry-run [--exclude relative/path] [--json]
--dry-run previews inventory and exclusions without uploading. --exclude is repeatable.
Omitted project is selected from a list. Later use send/take from any subfolder.
Use edda send FOLDER to create a new project and upload it in one guided workflow.`,
	"backup": `Create a verified backup on the machine holding the server data.
  edda backup [--db FILE] [--data ROOT] [--output NEW_DIRECTORY]
Missing paths are requested. This is a server administration command, not a download.`,
	"verify-backup": `Verify a backup's integrity.
  edda verify-backup [--source DIRECTORY]
The backup directory is requested if omitted.`,
	"restore-backup": `Restore a backup into a new data directory.
  edda restore-backup [--source DIRECTORY] [--output NEW_DATA_ROOT]
Missing paths are requested. Existing destinations are not overwritten.`,
	"init": `Initialize metadata for the earlier local snapshot workflow.
  edda init [FOLDER] [--title TITLE] [--id ID] [--server-url URL]
This does not upload or connect a folder. For server synchronization, use edda send.
The title is requested; ID is generated when omitted. Server URL is optional.`,
	"ids": `Maintain stable file IDs for a local prototype.
  edda ids sync [FOLDER]
With no subcommand, asks which action to perform. Not used in connected folders.`,
	"save": `Save a local prototype checkpoint or promote a file revision.
  edda save [FOLDER] NOTE
  edda save [FOLDER] --id FILE_ID (--from-draft | --body-file FILE) [--expected-sha256 HASH]
An omitted checkpoint note or required file choice is requested. This is a local
prototype operation; to upload files, use edda send FOLDER.`,
	"checkpoint": `Record a local prototype snapshot.
  edda checkpoint [FOLDER] [--message NOTE]
In a terminal an omitted note is requested. Nothing is uploaded.`,
	"files": `List files and stable IDs in a local prototype.
  edda files [FOLDER]
Not used in connected folders.`,
	"diff": `Compare a local prototype checkpoint with another checkpoint or current files.
  edda diff [FOLDER] [--from ID] [--to ID]
The source checkpoint is requested if omitted. Without --to, compare current files.`,
}

func printCommandHelp(command string, output io.Writer) error {
	text, ok := commandHelp[command]
	if !ok {
		return fmt.Errorf("unknown command %q", command)
	}
	fmt.Fprintf(output, "%s\n\nIn a terminal, missing required values are requested. Explicit arguments skip\nprompts. Optional settings keep their defaults. Connected project commands find .edda in\nparent directories. Ctrl+C cancels input.\n", text)
	return nil
}

func wantsCommandHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func prepareInteractivePath(command string, args []string, output io.Writer) ([]string, error) {
	switch command {
	case "send", "take", "status", "history", "restore", "conflicts", "resolve", "move":
		start, rest := splitOptionalPath(args)
		root, err := findProjectRoot(start)
		if err != nil {
			return nil, err
		}
		if root != "" {
			return append([]string{root}, rest...), nil
		}
	}

	if !interactiveInput() {
		return args, nil
	}
	for _, arg := range args {
		if arg == "--json" || strings.HasPrefix(arg, "--json=") {
			return args, nil
		}
	}
	switch command {
	case "save", "send", "attach", "get", "import", "status", "take", "move", "history", "restore", "conflicts", "resolve", "init", "checkpoint", "files", "diff":
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			return args, nil
		}
		var path string
		label, fallback := "Project folder", "."
		if command == "get" {
			label, fallback = "New destination folder", ""
		}
		if err := askValue(&path, label, fallback, output); err != nil {
			return nil, err
		}
		return append([]string{path}, args...), nil
	}
	return args, nil
}

// The nearest .edda is a boundary, even for a local prototype. Never silently
// skip broken/nested metadata and operate on a different project above it.
func findProjectRoot(start string) (string, error) {
	root, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		info, err := os.Lstat(filepath.Join(root, ".edda"))
		if err == nil {
			if !info.IsDir() {
				return "", fmt.Errorf("%s: .edda must be a real directory", root)
			}
			return root, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", nil
		}
		root = parent
	}
}
