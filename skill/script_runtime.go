package skill

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"git.inkyquill.net/inky/writer/skill/runtime"
	"git.inkyquill.net/inky/writer/store"
)

const (
	minScriptTimeoutMS          int64 = 100
	maxScriptTimeoutMS          int64 = 30000
	defaultScriptMaxStdoutBytes int64 = 65536
	defaultScriptMaxStderrBytes int64 = 16384
	minScriptMaxStreamBytes     int64 = 1024
)

func (s *Service) ListScriptAudits(ctx context.Context, projectID, skillID string) ([]ScriptAudit, error) {
	if strings.TrimSpace(projectID) == "" || strings.TrimSpace(skillID) == "" {
		return nil, ErrInvalidInput
	}
	rows, err := s.queries.ListSkillScriptAudits(ctx, store.ListSkillScriptAuditsParams{
		ProjectID: projectID,
		SkillID:   skillID,
	})
	if err != nil {
		return nil, fmt.Errorf("list script audits: %w", err)
	}
	audits := make([]ScriptAudit, 0, len(rows))
	for _, row := range rows {
		audit := scriptAuditFromStore(row)
		approval, err := s.queries.GetSkillScriptApprovalByFile(ctx, store.GetSkillScriptApprovalByFileParams{
			ProjectID:   projectID,
			SkillFileID: row.SkillFileID,
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get script approval: %w", err)
		}
		if err == nil {
			converted := scriptApprovalFromStore(approval)
			audit.Approval = &converted
		}
		audits = append(audits, audit)
	}
	return audits, nil
}

func (s *Service) UpdateScriptApproval(ctx context.Context, input UpdateScriptApprovalInput) (ScriptApproval, error) {
	if strings.TrimSpace(input.ProjectID) == "" || strings.TrimSpace(input.SkillID) == "" || strings.TrimSpace(input.SkillFileID) == "" {
		return ScriptApproval{}, ErrInvalidInput
	}
	input.RuntimeCommand = strings.TrimSpace(input.RuntimeCommand)
	if input.Enabled && input.RuntimeCommand == "" {
		return ScriptApproval{}, ErrInvalidInput
	}
	if input.TimeoutMS < minScriptTimeoutMS || input.TimeoutMS > maxScriptTimeoutMS {
		return ScriptApproval{}, ErrInvalidInput
	}
	input.MaxStdoutBytes = streamLimit(input.MaxStdoutBytes, defaultScriptMaxStdoutBytes)
	input.MaxStderrBytes = streamLimit(input.MaxStderrBytes, defaultScriptMaxStderrBytes)

	audit, err := s.queries.GetSkillScriptAuditByFile(ctx, store.GetSkillScriptAuditByFileParams{
		ProjectID:   input.ProjectID,
		SkillFileID: input.SkillFileID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ScriptApproval{}, ErrInvalidInput
		}
		return ScriptApproval{}, fmt.Errorf("get script audit: %w", err)
	}
	if audit.SkillID != input.SkillID {
		return ScriptApproval{}, ErrInvalidInput
	}
	if input.Enabled && (audit.DestructiveOperations != 0 || input.AllowNetwork || input.AllowProjectFiles) {
		return ScriptApproval{}, ErrInvalidInput
	}
	if !input.Enabled {
		input.AllowNetwork = false
		input.AllowProjectFiles = false
	}

	now := nowString()
	approvedBy := emptyDefault(input.ApprovedBy, "local-admin")
	if err := s.queries.UpsertSkillScriptApproval(ctx, store.UpsertSkillScriptApprovalParams{
		ID:                newID("scriptapproval"),
		ProjectID:         input.ProjectID,
		SkillID:           input.SkillID,
		SkillFileID:       input.SkillFileID,
		AuditID:           audit.ID,
		Enabled:           boolInt(input.Enabled),
		RuntimeCommand:    input.RuntimeCommand,
		TimeoutMs:         input.TimeoutMS,
		MaxStdoutBytes:    input.MaxStdoutBytes,
		MaxStderrBytes:    input.MaxStderrBytes,
		AllowNetwork:      boolInt(input.AllowNetwork),
		AllowProjectFiles: boolInt(input.AllowProjectFiles),
		ApprovedBy:        approvedBy,
		ApprovedAt:        now,
		UpdatedAt:         now,
	}); err != nil {
		return ScriptApproval{}, fmt.Errorf("upsert script approval: %w", err)
	}

	approval, err := s.queries.GetSkillScriptApprovalByFile(ctx, store.GetSkillScriptApprovalByFileParams{
		ProjectID:   input.ProjectID,
		SkillFileID: input.SkillFileID,
	})
	if err != nil {
		return ScriptApproval{}, fmt.Errorf("get script approval: %w", err)
	}
	return scriptApprovalFromStore(approval), nil
}

func (s *Service) RunScript(ctx context.Context, input RunScriptInput) (ScriptRun, error) {
	if strings.TrimSpace(input.ProjectID) == "" || strings.TrimSpace(input.SkillID) == "" {
		return ScriptRun{}, ErrInvalidInput
	}
	if strings.TrimSpace(input.SkillFileID) == "" && strings.TrimSpace(input.ScriptPath) == "" {
		return ScriptRun{}, ErrInvalidInput
	}

	project, err := s.queries.GetStoryProjectByID(ctx, input.ProjectID)
	if err != nil {
		return ScriptRun{}, fmt.Errorf("get story project: %w", err)
	}
	skillRow, err := s.queries.GetSkillByProjectID(ctx, store.GetSkillByProjectIDParams{
		ProjectID: input.ProjectID,
		ID:        input.SkillID,
	})
	if err != nil {
		return ScriptRun{}, fmt.Errorf("get skill: %w", err)
	}
	scriptFile, err := s.resolveScriptFile(ctx, input)
	if err != nil {
		return ScriptRun{}, err
	}
	if scriptFile.Purpose != string(FilePurposeScript) {
		return ScriptRun{}, ErrInvalidInput
	}

	approval, err := s.queries.GetSkillScriptApprovalByFile(ctx, store.GetSkillScriptApprovalByFileParams{
		ProjectID:   input.ProjectID,
		SkillFileID: scriptFile.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ScriptRun{}, ErrInvalidInput
		}
		return ScriptRun{}, fmt.Errorf("get script approval: %w", err)
	}
	if approval.SkillID != input.SkillID || approval.Enabled == 0 || strings.TrimSpace(approval.RuntimeCommand) == "" {
		return ScriptRun{}, ErrInvalidInput
	}
	audit, err := s.queries.GetSkillScriptAuditByFile(ctx, store.GetSkillScriptAuditByFileParams{
		ProjectID:   input.ProjectID,
		SkillFileID: scriptFile.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ScriptRun{}, ErrInvalidInput
		}
		return ScriptRun{}, fmt.Errorf("get script audit: %w", err)
	}
	if audit.SkillID != input.SkillID || audit.DestructiveOperations != 0 || approval.AllowNetwork != 0 || approval.AllowProjectFiles != 0 {
		return ScriptRun{}, ErrInvalidInput
	}

	envelope, err := s.buildScriptEnvelope(ctx, project, skillRow, scriptFile, input)
	if err != nil {
		return ScriptRun{}, err
	}
	request := runtime.RunRequest{
		Command:        approval.RuntimeCommand,
		Timeout:        time.Duration(approval.TimeoutMs) * time.Millisecond,
		MaxStdoutBytes: approval.MaxStdoutBytes,
		MaxStderrBytes: approval.MaxStderrBytes,
		Input:          envelope,
	}
	runner := s.scriptRunner
	if runner == nil {
		runner = runtime.NewRunner()
	}
	result, runErr := runner.Run(ctx, request)
	if runErr != nil {
		result = runtime.RunResult{
			Status:       runtime.StatusFailed,
			ErrorMessage: runErr.Error(),
		}
	}

	run, persistErr := s.persistScriptRun(ctx, input, scriptFile.ID, approval.ID, envelope, result)
	if persistErr != nil {
		return ScriptRun{}, persistErr
	}
	if runErr != nil {
		return run, fmt.Errorf("run script: %w", runErr)
	}
	return run, nil
}

func (s *Service) ListScriptRunsByProject(ctx context.Context, projectID string, limit int64) ([]ScriptRun, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, ErrInvalidInput
	}
	rows, err := s.queries.ListSkillScriptRunsByProject(ctx, store.ListSkillScriptRunsByProjectParams{
		ProjectID: projectID,
		Limit:     normalizedLimit(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list script runs by project: %w", err)
	}
	return scriptRunsFromStore(rows), nil
}

func (s *Service) ListScriptRunsBySession(ctx context.Context, projectID, sessionID string, limit int64) ([]ScriptRun, error) {
	if strings.TrimSpace(projectID) == "" || strings.TrimSpace(sessionID) == "" {
		return nil, ErrInvalidInput
	}
	rows, err := s.queries.ListSkillScriptRunsBySession(ctx, store.ListSkillScriptRunsBySessionParams{
		ProjectID: projectID,
		SessionID: sql.NullString{String: sessionID, Valid: true},
		Limit:     normalizedLimit(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list script runs by session: %w", err)
	}
	return scriptRunsFromStore(rows), nil
}

func (s *Service) resolveScriptFile(ctx context.Context, input RunScriptInput) (store.SkillFile, error) {
	if strings.TrimSpace(input.ScriptPath) != "" {
		file, err := s.queries.GetSkillFile(ctx, store.GetSkillFileParams{
			ProjectID:    input.ProjectID,
			SkillID:      input.SkillID,
			RelativePath: input.ScriptPath,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return store.SkillFile{}, ErrInvalidInput
			}
			return store.SkillFile{}, fmt.Errorf("get script file: %w", err)
		}
		if strings.TrimSpace(input.SkillFileID) != "" && file.ID != input.SkillFileID {
			return store.SkillFile{}, ErrInvalidInput
		}
		return file, nil
	}

	files, err := s.queries.ListSkillFiles(ctx, input.SkillID)
	if err != nil {
		return store.SkillFile{}, fmt.Errorf("list skill files: %w", err)
	}
	for _, file := range files {
		if file.ID == input.SkillFileID {
			return file, nil
		}
	}
	return store.SkillFile{}, ErrInvalidInput
}

func (s *Service) buildScriptEnvelope(ctx context.Context, project store.StoryProject, skillRow store.Skill, scriptFile store.SkillFile, input RunScriptInput) (runtime.Envelope, error) {
	envelope := runtime.Envelope{
		RuntimeVersion: runtime.RuntimeVersion,
		Project: runtime.EnvelopeRef{
			ID:       project.ID,
			Title:    project.Title,
			Language: project.Language,
		},
		Skill: runtime.EnvelopeRef{
			ID:    skillRow.ID,
			Name:  skillRow.Name,
			Title: skillRow.DisplayName,
		},
		Script: runtime.ScriptRef{
			FileID:       scriptFile.ID,
			RelativePath: scriptFile.RelativePath,
		},
		Inputs: runtime.EnvelopeInputs{
			ContentIDs: dedupeStrings(input.ContentIDs),
			Arguments:  input.Arguments,
		},
	}
	for _, contentID := range envelope.Inputs.ContentIDs {
		if _, err := s.queries.GetContentItem(ctx, store.GetContentItemParams{ID: contentID, ProjectID: input.ProjectID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return runtime.Envelope{}, ErrInvalidInput
			}
			return runtime.Envelope{}, fmt.Errorf("get content item: %w", err)
		}
	}
	for _, section := range input.EntrySections {
		contentID := strings.TrimSpace(section.ContentID)
		heading := strings.TrimSpace(section.Heading)
		if contentID == "" || heading == "" {
			return runtime.Envelope{}, ErrInvalidInput
		}
		if err := s.validateEntrySection(ctx, input.ProjectID, contentID, heading); err != nil {
			return runtime.Envelope{}, err
		}
		envelope.Inputs.EntrySections = append(envelope.Inputs.EntrySections, runtime.EntrySectionRef{
			ContentID: contentID,
			Heading:   heading,
		})
	}
	for _, assetPath := range dedupeStrings(input.AssetPaths) {
		file, err := s.queries.GetSkillFile(ctx, store.GetSkillFileParams{
			ProjectID:    input.ProjectID,
			SkillID:      input.SkillID,
			RelativePath: assetPath,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return runtime.Envelope{}, ErrInvalidInput
			}
			return runtime.Envelope{}, fmt.Errorf("get requested asset: %w", err)
		}
		if file.Purpose == string(FilePurposeScript) {
			return runtime.Envelope{}, ErrInvalidInput
		}
		envelope.Inputs.Assets = append(envelope.Inputs.Assets, runtime.AssetInput{
			Path:     file.RelativePath,
			BodyText: file.BodyText,
		})
	}
	return envelope, nil
}

func (s *Service) validateEntrySection(ctx context.Context, projectID, contentID, heading string) error {
	sections, err := s.queries.ListEntrySections(ctx, store.ListEntrySectionsParams{
		ContentItemID: contentID,
		ProjectID:     projectID,
	})
	if err != nil {
		return fmt.Errorf("list entry sections: %w", err)
	}
	for _, section := range sections {
		if section.Heading == heading {
			return nil
		}
	}
	return ErrInvalidInput
}

func (s *Service) persistScriptRun(ctx context.Context, input RunScriptInput, skillFileID, approvalID string, envelope runtime.Envelope, result runtime.RunResult) (ScriptRun, error) {
	now := nowString()
	inputJSON, err := marshalJSON(envelope)
	if err != nil {
		return ScriptRun{}, err
	}
	outputJSON := "{}"
	outputKind := string(result.Output.Kind)
	if outputKind != "" {
		outputJSON, err = marshalJSON(result.Output)
		if err != nil {
			return ScriptRun{}, err
		}
	}
	status := result.Status
	if status == "" {
		status = runtime.StatusFailed
	}
	params := store.CreateSkillScriptRunParams{
		ID:           newID("scriptrun"),
		ProjectID:    input.ProjectID,
		SessionID:    sql.NullString{String: input.SessionID, Valid: strings.TrimSpace(input.SessionID) != ""},
		SkillID:      input.SkillID,
		SkillFileID:  skillFileID,
		ApprovalID:   approvalID,
		ToolCallID:   input.ToolCallID,
		Status:       string(status),
		OutputKind:   outputKind,
		InputJson:    inputJSON,
		OutputJson:   outputJSON,
		StdoutText:   result.StdoutText,
		StderrText:   result.StderrText,
		ExitCode:     int64(result.ExitCode),
		DurationMs:   result.Duration.Milliseconds(),
		ErrorMessage: result.ErrorMessage,
		CreatedAt:    now,
	}
	if err := s.queries.CreateSkillScriptRun(ctx, params); err != nil {
		return ScriptRun{}, fmt.Errorf("create script run: %w", err)
	}
	return scriptRunFromStore(store.SkillScriptRun{
		ID:           params.ID,
		ProjectID:    params.ProjectID,
		SessionID:    params.SessionID,
		SkillID:      params.SkillID,
		SkillFileID:  params.SkillFileID,
		ApprovalID:   params.ApprovalID,
		ToolCallID:   params.ToolCallID,
		Status:       params.Status,
		OutputKind:   params.OutputKind,
		InputJson:    params.InputJson,
		OutputJson:   params.OutputJson,
		StdoutText:   params.StdoutText,
		StderrText:   params.StderrText,
		ExitCode:     params.ExitCode,
		DurationMs:   params.DurationMs,
		ErrorMessage: params.ErrorMessage,
		CreatedAt:    params.CreatedAt,
	}), nil
}

func runtimeFromScriptPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".py":
		return "python"
	case ".js", ".mjs", ".cjs":
		return "node"
	case ".ts":
		return "typescript"
	case ".sh", ".bash", ".zsh":
		return "shell"
	default:
		return "unknown"
	}
}

func defaultScriptExpectedInputsJSON() string {
	return `{"contentIds":"optional","entrySections":"optional","assets":"optional","arguments":"optional"}`
}

func defaultScriptExpectedOutputsJSON() string {
	return `{"contract":"skill-script-output/v1","kinds":["report","proposal","draft","generated_data"]}`
}

func scriptAuditFromStore(row store.SkillScriptAudit) ScriptAudit {
	return ScriptAudit{
		ID:                    row.ID,
		ProjectID:             row.ProjectID,
		SkillID:               row.SkillID,
		SkillFileID:           row.SkillFileID,
		RelativePath:          row.RelativePath,
		Runtime:               row.Runtime,
		DestructiveOperations: row.DestructiveOperations != 0,
		FilesystemAccess:      row.FilesystemAccess,
		NetworkAccess:         row.NetworkAccess != 0,
		ExternalDependencies:  row.ExternalDependencies,
		ExpectedInputsJSON:    row.ExpectedInputsJson,
		ExpectedOutputsJSON:   row.ExpectedOutputsJson,
		RiskNotes:             row.RiskNotes,
		Recommendation:        ScriptRecommendation(row.Recommendation),
		AuditedAt:             row.AuditedAt,
	}
}

func scriptApprovalFromStore(row store.SkillScriptApproval) ScriptApproval {
	return ScriptApproval{
		ID:                row.ID,
		ProjectID:         row.ProjectID,
		SkillID:           row.SkillID,
		SkillFileID:       row.SkillFileID,
		AuditID:           row.AuditID,
		Enabled:           row.Enabled != 0,
		RuntimeCommand:    row.RuntimeCommand,
		TimeoutMS:         row.TimeoutMs,
		MaxStdoutBytes:    row.MaxStdoutBytes,
		MaxStderrBytes:    row.MaxStderrBytes,
		AllowNetwork:      row.AllowNetwork != 0,
		AllowProjectFiles: row.AllowProjectFiles != 0,
		ApprovedBy:        row.ApprovedBy,
		ApprovedAt:        row.ApprovedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

func scriptRunsFromStore(rows []store.SkillScriptRun) []ScriptRun {
	runs := make([]ScriptRun, 0, len(rows))
	for _, row := range rows {
		runs = append(runs, scriptRunFromStore(row))
	}
	return runs
}

func scriptRunFromStore(row store.SkillScriptRun) ScriptRun {
	run := ScriptRun{
		ID:           row.ID,
		ProjectID:    row.ProjectID,
		SkillID:      row.SkillID,
		SkillFileID:  row.SkillFileID,
		ApprovalID:   row.ApprovalID,
		ToolCallID:   row.ToolCallID,
		Status:       ScriptRunStatus(row.Status),
		OutputKind:   row.OutputKind,
		InputJSON:    row.InputJson,
		OutputJSON:   row.OutputJson,
		StdoutText:   row.StdoutText,
		StderrText:   row.StderrText,
		ExitCode:     row.ExitCode,
		DurationMS:   row.DurationMs,
		ErrorMessage: row.ErrorMessage,
		CreatedAt:    row.CreatedAt,
	}
	if row.SessionID.Valid {
		run.SessionID = row.SessionID.String
	}
	if row.OutputJson != "" && row.OutputJson != "{}" {
		_ = json.Unmarshal([]byte(row.OutputJson), &run.Output)
	}
	return run
}

func dedupeStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizedLimit(limit int64) int64 {
	if limit <= 0 {
		return 50
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func streamLimit(value, fallback int64) int64 {
	if value <= 0 {
		return fallback
	}
	if value < minScriptMaxStreamBytes {
		return minScriptMaxStreamBytes
	}
	return value
}
