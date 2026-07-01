# Linear Checkpoints Instead of Git

Open Edda provides simple linear checkpoints for project history instead of exposing git concepts. A checkpoint is a named project-wide snapshot of saved files and syncable `.edda/` metadata. Authors use checkpoints to compare changes, restore earlier states, recover from mistakes, and move saved work between a local folder and the server.

The product should use writing-oriented commands and labels, such as `edda save "Note"`, `edda history`, `edda diff`, and `edda restore`. It should not require branches, staging, rebases, remotes, merge commits, or git terminology in the main workflow.

The first implementation should favor reliability over storage efficiency. Full snapshots or simple content-addressed blobs are acceptable until real project histories show that delta storage is necessary.
