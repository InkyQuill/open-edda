# Skill Scripts Require Admin Approval

Open Edda imports skill instructions, routing metadata, templates, data files, and script files as usable agent context, but imported scripts are disabled by default. Script execution requires an explicit admin approval record for the individual script file.

Approved scripts run through the Skill Script Runtime. Open Edda supplies database-backed JSON inputs and an empty temporary working directory, then accepts only structured JSON outputs: reports, proposals, generated data, or drafts. Scripts must not directly mutate Story Text, Story Bible Entries, Entry Sections, Project Notes, Attached Notes, or project structure.

The first runtime rejects network-enabled and project-file-enabled approvals. If a script needs those capabilities, it remains disabled until a later runtime policy can provide a stronger sandbox and review model.
