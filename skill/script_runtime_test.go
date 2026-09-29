package skill

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"git.inkyquill.net/inky/writer/skill/runtime"
)

type fakeScriptRunner struct {
	requests []runtime.RunRequest
	result   runtime.RunResult
	err      error
}

func (r *fakeScriptRunner) Run(ctx context.Context, request runtime.RunRequest) (runtime.RunResult, error) {
	r.requests = append(r.requests, request)
	return r.result, r.err
}

func TestInstallCreatesDisabledScriptAudit(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	installed, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported:    importedStylePass("Rewrite template"),
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	audits, err := service.ListScriptAudits(ctx, "project-1", installed.ID)
	if err != nil {
		t.Fatalf("ListScriptAudits() error = %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("audit count = %d, want 1", len(audits))
	}
	audit := audits[0]
	if audit.SkillFileID == "" {
		t.Fatal("audit SkillFileID is empty")
	}
	if audit.SkillFileID != scriptFileID(t, installed) {
		t.Fatalf("audit SkillFileID = %q, want created script file ID", audit.SkillFileID)
	}
	if audit.RelativePath != "scripts/analyze.sh" {
		t.Fatalf("audit RelativePath = %q", audit.RelativePath)
	}
	if audit.Runtime != "shell" {
		t.Fatalf("audit Runtime = %q, want shell", audit.Runtime)
	}
	if audit.Recommendation != ScriptRecommendationDisabled {
		t.Fatalf("audit Recommendation = %q, want disabled", audit.Recommendation)
	}
	if audit.DestructiveOperations || audit.NetworkAccess {
		t.Fatalf("audit safe defaults = destructive %v network %v, want false", audit.DestructiveOperations, audit.NetworkAccess)
	}
	if audit.FilesystemAccess != "temp_workspace" {
		t.Fatalf("audit FilesystemAccess = %q, want temp_workspace", audit.FilesystemAccess)
	}
	var outputContract map[string]any
	if err := json.Unmarshal([]byte(audit.ExpectedOutputsJSON), &outputContract); err != nil {
		t.Fatalf("unmarshal expected output contract: %v", err)
	}
	if outputContract["contract"] != "skill-script-output/v1" {
		t.Fatalf("output contract = %#v, want skill-script-output/v1", outputContract)
	}
}

func TestRunScriptRequiresApproval(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	service.SetScriptRunner(&fakeScriptRunner{})
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)

	_, err := service.RunScript(ctx, RunScriptInput{
		ProjectID:     "project-1",
		SkillID:       installed.ID,
		SkillFileID:   scriptFileID(t, installed),
		ToolCallID:    "tool-call-1",
		ContentIDs:    []string{"content-1"},
		Arguments:     map[string]any{"mode": "report"},
		EntrySections: []ScriptEntrySectionInput{{ContentID: "content-1", Heading: "Opening"}},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("RunScript() error = %v, want ErrInvalidInput", err)
	}
}

func TestRunScriptPersistsReportRun(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	runner := &fakeScriptRunner{result: runtime.RunResult{
		Status: runtime.StatusSucceeded,
		Output: runtime.ScriptOutput{
			Kind:     runtime.OutputKindReport,
			Title:    "Style report",
			Markdown: "Tighten two sentences.",
		},
		StdoutText: `{"kind":"report","title":"Style report","markdown":"Tighten two sentences."}`,
		StderrText: "note",
		ExitCode:   0,
		Duration:   12 * time.Millisecond,
	}}
	service.SetScriptRunner(runner)
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	seedScriptRuntimeContent(t, db)
	approval := approveRuntimeScript(t, service, installed, false, false)

	run, err := service.RunScript(ctx, RunScriptInput{
		ProjectID:     "project-1",
		SessionID:     "session-1",
		SkillID:       installed.ID,
		ScriptPath:    "scripts/analyze.sh",
		ToolCallID:    "tool-call-1",
		ContentIDs:    []string{"content-1"},
		AssetPaths:    []string{"templates/rewrite.md"},
		Arguments:     map[string]any{"mode": "report"},
		EntrySections: []ScriptEntrySectionInput{{ContentID: "content-1", Heading: "Opening"}},
	})
	if err != nil {
		t.Fatalf("RunScript() error = %v", err)
	}
	if run.Status != ScriptRunStatusSucceeded || run.OutputKind != string(runtime.OutputKindReport) {
		t.Fatalf("run status/kind = %q/%q, want succeeded/report", run.Status, run.OutputKind)
	}
	if run.ApprovalID != approval.ID {
		t.Fatalf("run ApprovalID = %q, want %q", run.ApprovalID, approval.ID)
	}
	if len(runner.requests) != 1 {
		t.Fatalf("runner requests = %d, want 1", len(runner.requests))
	}
	request := runner.requests[0]
	if request.Command != "sh scripts/analyze.sh" {
		t.Fatalf("request Command = %q", request.Command)
	}
	if request.Input.RuntimeVersion != runtime.RuntimeVersion {
		t.Fatalf("runtime version = %q", request.Input.RuntimeVersion)
	}
	if request.Input.Project.ID != "project-1" || request.Input.Project.Title != "Test" || request.Input.Project.Language != "en" {
		t.Fatalf("project envelope = %#v", request.Input.Project)
	}
	if request.Input.Skill.ID != installed.ID || request.Input.Skill.Name != "style-pass" {
		t.Fatalf("skill envelope = %#v", request.Input.Skill)
	}
	if request.Input.Script.FileID != scriptFileID(t, installed) || request.Input.Script.RelativePath != "scripts/analyze.sh" {
		t.Fatalf("script envelope = %#v", request.Input.Script)
	}
	if len(request.Input.Inputs.ContentItems) != 1 || request.Input.Inputs.ContentItems[0].ID != "content-1" || request.Input.Inputs.ContentItems[0].BodyMarkdown != "Original body" || request.Input.Inputs.ContentItems[0].Kind != "chapter" {
		t.Fatalf("content items = %#v", request.Input.Inputs.ContentItems)
	}
	if len(request.Input.Inputs.EntrySections) != 1 || request.Input.Inputs.EntrySections[0].Heading != "Opening" || request.Input.Inputs.EntrySections[0].BodyMarkdown != "Original body" {
		t.Fatalf("entry sections = %#v", request.Input.Inputs.EntrySections)
	}
	if len(request.Input.Inputs.Assets) != 1 || request.Input.Inputs.Assets[0].Path != "templates/rewrite.md" || request.Input.Inputs.Assets[0].BodyText == "" {
		t.Fatalf("assets = %#v", request.Input.Inputs.Assets)
	}

	runs, err := service.ListScriptRunsByProject(ctx, "project-1", 10)
	if err != nil {
		t.Fatalf("ListScriptRunsByProject() error = %v", err)
	}
	if len(runs) != 1 || runs[0].ID != run.ID || runs[0].StdoutText == "" {
		t.Fatalf("project runs = %#v, want persisted run", runs)
	}
	sessionRuns, err := service.ListScriptRunsBySession(ctx, "project-1", "session-1", 10)
	if err != nil {
		t.Fatalf("ListScriptRunsBySession() error = %v", err)
	}
	if len(sessionRuns) != 1 || sessionRuns[0].ID != run.ID {
		t.Fatalf("session runs = %#v, want persisted run", sessionRuns)
	}
}

func TestListScriptAuditsIncludesApproval(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	approval := approveRuntimeScript(t, service, installed, false, false)

	audits, err := service.ListScriptAudits(ctx, "project-1", installed.ID)
	if err != nil {
		t.Fatalf("ListScriptAudits() error = %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("audit count = %d, want 1", len(audits))
	}
	if audits[0].Approval == nil {
		t.Fatal("audit Approval is nil, want matching approval")
	}
	if audits[0].Approval.ID != approval.ID || audits[0].Approval.SkillFileID != scriptFileID(t, installed) {
		t.Fatalf("audit Approval = %#v, want approval %#v", audits[0].Approval, approval)
	}
}

func TestRunScriptPersistsRejectedRuntimeResult(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	service.SetScriptRunner(&fakeScriptRunner{result: runtime.RunResult{
		Status:       runtime.StatusRejected,
		StdoutText:   "not-json",
		StderrText:   "bad output",
		ExitCode:     0,
		Duration:     3 * time.Millisecond,
		ErrorMessage: "script stdout must be valid JSON",
	}})
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	approveRuntimeScript(t, service, installed, false, false)

	run, err := service.RunScript(ctx, RunScriptInput{
		ProjectID:   "project-1",
		SkillID:     installed.ID,
		SkillFileID: scriptFileID(t, installed),
	})
	if err != nil {
		t.Fatalf("RunScript() error = %v", err)
	}
	if run.Status != ScriptRunStatusRejected || run.ErrorMessage != "script stdout must be valid JSON" {
		t.Fatalf("run = %#v, want persisted rejected result", run)
	}
	runs, err := service.ListScriptRunsByProject(ctx, "project-1", 10)
	if err != nil {
		t.Fatalf("ListScriptRunsByProject() error = %v", err)
	}
	if len(runs) != 1 || runs[0].Status != ScriptRunStatusRejected {
		t.Fatalf("runs = %#v, want persisted rejected run", runs)
	}
}

func TestRunScriptPersistsFailedRuntimeResult(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	service.SetScriptRunner(&fakeScriptRunner{
		result: runtime.RunResult{
			Status:       runtime.StatusFailed,
			StdoutText:   "",
			StderrText:   "boom",
			ExitCode:     1,
			Duration:     4 * time.Millisecond,
			ErrorMessage: "script exited with status 1",
		},
	})
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	approveRuntimeScript(t, service, installed, false, false)

	run, err := service.RunScript(ctx, RunScriptInput{
		ProjectID:   "project-1",
		SkillID:     installed.ID,
		SkillFileID: scriptFileID(t, installed),
	})
	if err != nil {
		t.Fatalf("RunScript() error = %v", err)
	}
	if run.Status != ScriptRunStatusFailed || run.ErrorMessage != "script exited with status 1" {
		t.Fatalf("run = %#v, want persisted failed result", run)
	}
	runs, err := service.ListScriptRunsByProject(ctx, "project-1", 10)
	if err != nil {
		t.Fatalf("ListScriptRunsByProject() error = %v", err)
	}
	if len(runs) != 1 || runs[0].Status != ScriptRunStatusFailed {
		t.Fatalf("runs = %#v, want persisted failed run", runs)
	}
}

func TestRunScriptRejectsUnsafeStoredApprovalOrAudit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(t *testing.T, db interface {
			Exec(query string, args ...any) (sql.Result, error)
		}, skillFileID string)
	}{
		{
			name: "network approval",
			mutate: func(t *testing.T, db interface {
				Exec(query string, args ...any) (sql.Result, error)
			}, skillFileID string) {
				t.Helper()
				if _, err := db.Exec(`UPDATE skill_script_approvals SET allow_network = 1 WHERE skill_file_id = ?`, skillFileID); err != nil {
					t.Fatalf("mutate approval: %v", err)
				}
			},
		},
		{
			name: "project files approval",
			mutate: func(t *testing.T, db interface {
				Exec(query string, args ...any) (sql.Result, error)
			}, skillFileID string) {
				t.Helper()
				if _, err := db.Exec(`UPDATE skill_script_approvals SET allow_project_files = 1 WHERE skill_file_id = ?`, skillFileID); err != nil {
					t.Fatalf("mutate approval: %v", err)
				}
			},
		},
		{
			name: "destructive audit",
			mutate: func(t *testing.T, db interface {
				Exec(query string, args ...any) (sql.Result, error)
			}, skillFileID string) {
				t.Helper()
				if _, err := db.Exec(`UPDATE skill_script_audits SET destructive_operations = 1 WHERE skill_file_id = ?`, skillFileID); err != nil {
					t.Fatalf("mutate audit: %v", err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := openMigratedTestDB(t)
			service := NewService(db)
			runner := &fakeScriptRunner{result: runtime.RunResult{Status: runtime.StatusSucceeded}}
			service.SetScriptRunner(runner)
			ctx := context.Background()

			installed := installRuntimeSkill(t, service)
			scriptID := scriptFileID(t, installed)
			approveRuntimeScript(t, service, installed, false, false)
			tc.mutate(t, db, scriptID)

			_, err := service.RunScript(ctx, RunScriptInput{
				ProjectID:   "project-1",
				SkillID:     installed.ID,
				SkillFileID: scriptID,
			})
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("RunScript() error = %v, want ErrInvalidInput", err)
			}
			if len(runner.requests) != 0 {
				t.Fatalf("runner requests = %d, want 0", len(runner.requests))
			}
		})
	}
}

func TestUpdateScriptApprovalRejectsNetworkAndProjectFiles(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	for _, tc := range []struct {
		name              string
		allowNetwork      bool
		allowProjectFiles bool
	}{
		{name: "network", allowNetwork: true},
		{name: "project files", allowProjectFiles: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.UpdateScriptApproval(ctx, UpdateScriptApprovalInput{
				ProjectID:         "project-1",
				SkillID:           installed.ID,
				SkillFileID:       scriptFileID(t, installed),
				Enabled:           true,
				RuntimeCommand:    "sh scripts/analyze.sh",
				TimeoutMS:         5000,
				MaxStdoutBytes:    65536,
				MaxStderrBytes:    16384,
				AllowNetwork:      tc.allowNetwork,
				AllowProjectFiles: tc.allowProjectFiles,
				ApprovedBy:        "tester",
			})
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("UpdateScriptApproval() error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestUpdateScriptApprovalAllowsDisablingUnsafeStoredApproval(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	scriptID := scriptFileID(t, installed)
	approveRuntimeScript(t, service, installed, false, false)
	if _, err := db.Exec(`UPDATE skill_script_approvals SET allow_network = 1, allow_project_files = 1 WHERE skill_file_id = ?`, scriptID); err != nil {
		t.Fatalf("mutate approval: %v", err)
	}
	if _, err := db.Exec(`UPDATE skill_script_audits SET destructive_operations = 1 WHERE skill_file_id = ?`, scriptID); err != nil {
		t.Fatalf("mutate audit: %v", err)
	}

	approval, err := service.UpdateScriptApproval(ctx, UpdateScriptApprovalInput{
		ProjectID:         "project-1",
		SkillID:           installed.ID,
		SkillFileID:       scriptID,
		Enabled:           false,
		RuntimeCommand:    "",
		TimeoutMS:         5000,
		AllowNetwork:      true,
		AllowProjectFiles: true,
		ApprovedBy:        "tester",
	})
	if err != nil {
		t.Fatalf("UpdateScriptApproval() error = %v", err)
	}
	if approval.Enabled || approval.AllowNetwork || approval.AllowProjectFiles {
		t.Fatalf("approval = %#v, want disabled safe policy", approval)
	}
}

func TestUpdateScriptApprovalRejectsTimeoutOutsideRange(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	for _, timeoutMS := range []int64{0, 99, 30001} {
		_, err := service.UpdateScriptApproval(ctx, UpdateScriptApprovalInput{
			ProjectID:      "project-1",
			SkillID:        installed.ID,
			SkillFileID:    scriptFileID(t, installed),
			Enabled:        true,
			RuntimeCommand: "sh scripts/analyze.sh",
			TimeoutMS:      timeoutMS,
			ApprovedBy:     "tester",
		})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("UpdateScriptApproval(timeout %d) error = %v, want ErrInvalidInput", timeoutMS, err)
		}
	}
}

func installRuntimeSkill(t *testing.T, service *Service) Skill {
	t.Helper()

	ctx := context.Background()
	installed, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported:    importedStylePass("Rewrite template"),
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	return installed
}

func approveRuntimeScript(t *testing.T, service *Service, installed Skill, allowNetwork, allowProjectFiles bool) ScriptApproval {
	t.Helper()

	approval, err := service.UpdateScriptApproval(context.Background(), UpdateScriptApprovalInput{
		ProjectID:         "project-1",
		SkillID:           installed.ID,
		SkillFileID:       scriptFileID(t, installed),
		Enabled:           true,
		RuntimeCommand:    "sh scripts/analyze.sh",
		TimeoutMS:         5000,
		MaxStdoutBytes:    65536,
		MaxStderrBytes:    16384,
		AllowNetwork:      allowNetwork,
		AllowProjectFiles: allowProjectFiles,
		ApprovedBy:        "tester",
	})
	if err != nil {
		t.Fatalf("UpdateScriptApproval() error = %v", err)
	}
	return approval
}

func scriptFileID(t *testing.T, installed Skill) string {
	t.Helper()

	for _, file := range installed.Files {
		if file.Purpose == FilePurposeScript {
			return file.ID
		}
	}
	t.Fatalf("installed skill has no script file: %#v", installed.Files)
	return ""
}

func seedScriptRuntimeContent(t *testing.T, db interface {
	Exec(query string, args ...any) (sql.Result, error)
}) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO content_items (
			id, project_id, kind, title, slug, body_markdown, metadata_json,
			sort_order, current_revision, created_at, updated_at
		) VALUES (
			'content-1', 'project-1', 'chapter', 'Opening', 'opening', 'Original body',
			'{}', 1, 1, '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
		);
		INSERT INTO entry_sections (id, content_item_id, heading, body_markdown, sort_order)
		VALUES ('section-1', 'content-1', 'Opening', 'Original body', 1);
	`)
	if err != nil {
		t.Fatalf("seed script runtime content: %v", err)
	}
}

func TestBuildEnvelopePopulatesContentItems(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	runner := &fakeScriptRunner{result: runtime.RunResult{
		Status: runtime.StatusSucceeded,
		Output: runtime.ScriptOutput{
			Kind:     runtime.OutputKindReport,
			Title:    "Test report",
			Markdown: "ok",
		},
		StdoutText: `{"kind":"report","title":"Test report","markdown":"ok"}`,
		ExitCode:   0,
		Duration:   1 * time.Millisecond,
	}}
	service.SetScriptRunner(runner)
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	seedScriptRuntimeContent(t, db)
	approveRuntimeScript(t, service, installed, false, false)

	run, err := service.RunScript(ctx, RunScriptInput{
		ProjectID:  "project-1",
		SessionID:  "session-1",
		SkillID:    installed.ID,
		ScriptPath: "scripts/analyze.sh",
		ToolCallID: "tool-call-content-items",
		ContentIDs: []string{"content-1"},
	})
	if err != nil {
		t.Fatalf("RunScript() error = %v", err)
	}
	if run.Status != ScriptRunStatusSucceeded {
		t.Fatalf("run status = %q, want succeeded", run.Status)
	}
	if len(runner.requests) != 1 {
		t.Fatalf("runner requests = %d, want 1", len(runner.requests))
	}
	input := runner.requests[0].Input
	if len(input.Inputs.ContentIDs) != 1 || input.Inputs.ContentIDs[0] != "content-1" {
		t.Fatalf("content IDs = %#v", input.Inputs.ContentIDs)
	}
	if len(input.Inputs.ContentItems) != 1 {
		t.Fatalf("content items = %#v, want 1 item", input.Inputs.ContentItems)
	}
	item := input.Inputs.ContentItems[0]
	if item.ID != "content-1" {
		t.Fatalf("content item ID = %q, want content-1", item.ID)
	}
	if item.Kind != "chapter" {
		t.Fatalf("content item kind = %q, want chapter", item.Kind)
	}
	if item.Title != "Opening" {
		t.Fatalf("content item title = %q, want Opening", item.Title)
	}
	if item.BodyMarkdown != "Original body" {
		t.Fatalf("content item body = %q, want 'Original body'", item.BodyMarkdown)
	}
	if item.Truncated {
		t.Fatal("content item should not be truncated")
	}
}

func TestBuildEnvelopePopulatesEntrySectionBody(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	runner := &fakeScriptRunner{result: runtime.RunResult{
		Status: runtime.StatusSucceeded,
		Output: runtime.ScriptOutput{
			Kind:     runtime.OutputKindReport,
			Title:    "Test report",
			Markdown: "ok",
		},
		StdoutText: `{"kind":"report","title":"Test report","markdown":"ok"}`,
		ExitCode:   0,
		Duration:   1 * time.Millisecond,
	}}
	service.SetScriptRunner(runner)
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	seedScriptRuntimeContent(t, db)
	approveRuntimeScript(t, service, installed, false, false)

	run, err := service.RunScript(ctx, RunScriptInput{
		ProjectID:  "project-1",
		SessionID:  "session-1",
		SkillID:    installed.ID,
		ScriptPath: "scripts/analyze.sh",
		ToolCallID: "tool-call-section-body",
		EntrySections: []ScriptEntrySectionInput{{
			ContentID: "content-1",
			Heading:   "Opening",
		}},
	})
	if err != nil {
		t.Fatalf("RunScript() error = %v", err)
	}
	if run.Status != ScriptRunStatusSucceeded {
		t.Fatalf("run status = %q, want succeeded", run.Status)
	}
	input := runner.requests[0].Input
	if len(input.Inputs.EntrySections) != 1 {
		t.Fatalf("entry sections = %#v, want 1", input.Inputs.EntrySections)
	}
	section := input.Inputs.EntrySections[0]
	if section.BodyMarkdown != "Original body" {
		t.Fatalf("entry section body = %q, want 'Original body'", section.BodyMarkdown)
	}
}

func TestBuildEnvelopeRejectsUnknownContentID(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	runner := &fakeScriptRunner{result: runtime.RunResult{Status: runtime.StatusSucceeded}}
	service.SetScriptRunner(runner)
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)
	approveRuntimeScript(t, service, installed, false, false)

	_, err := service.RunScript(ctx, RunScriptInput{
		ProjectID:  "project-1",
		SkillID:    installed.ID,
		ScriptPath: "scripts/analyze.sh",
		ToolCallID: "tool-call-missing-content",
		ContentIDs: []string{"nonexistent-content"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("RunScript() error = %v, want ErrInvalidInput", err)
	}
}

func TestBuildEnvelopeTruncatesLargeContentBody(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	runner := &fakeScriptRunner{result: runtime.RunResult{
		Status: runtime.StatusSucceeded,
		Output: runtime.ScriptOutput{
			Kind:     runtime.OutputKindReport,
			Title:    "Test report",
			Markdown: "ok",
		},
		StdoutText: `{"kind":"report","title":"Test report","markdown":"ok"}`,
		ExitCode:   0,
		Duration:   1 * time.Millisecond,
	}}
	service.SetScriptRunner(runner)
	ctx := context.Background()

	installed := installRuntimeSkill(t, service)

	largeBody := strings.Repeat("a", int(maxContentBodyBytes)+1000)
	_, err := db.Exec(`
		INSERT INTO content_items (
			id, project_id, kind, title, slug, body_markdown, metadata_json,
			sort_order, current_revision, created_at, updated_at
		) VALUES (
			'content-large', 'project-1', 'chapter', 'Big Chapter', 'big-chapter', ?, '{}',
			1, 1, '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
		);
	`, largeBody)
	if err != nil {
		t.Fatalf("seed large content: %v", err)
	}

	approveRuntimeScript(t, service, installed, false, false)

	run, err := service.RunScript(ctx, RunScriptInput{
		ProjectID:  "project-1",
		SessionID:  "session-1",
		SkillID:    installed.ID,
		ScriptPath: "scripts/analyze.sh",
		ToolCallID: "tool-call-truncation",
		ContentIDs: []string{"content-large"},
	})
	if err != nil {
		t.Fatalf("RunScript() error = %v", err)
	}
	if run.Status != ScriptRunStatusSucceeded {
		t.Fatalf("run status = %q, want succeeded", run.Status)
	}
	input := runner.requests[0].Input
	if len(input.Inputs.ContentItems) != 1 {
		t.Fatalf("content items = %#v, want 1", input.Inputs.ContentItems)
	}
	item := input.Inputs.ContentItems[0]
	if !item.Truncated {
		t.Fatal("large content item should be truncated")
	}
	if int64(len(item.BodyMarkdown)) > maxContentBodyBytes+int64(len(contentTruncationMarker))+1 {
		t.Fatalf("truncated body length = %d, expected near %d", len(item.BodyMarkdown), maxContentBodyBytes)
	}
}
