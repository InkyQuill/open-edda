package skill

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"sync/atomic"
	"time"

	"git.inkyquill.net/inky/writer/store"
)

var ErrInvalidInput = errors.New("invalid skill input")

type Service struct {
	db      *sql.DB
	queries *store.Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db, queries: store.New(db)}
}

func (s *Service) Install(ctx context.Context, input InstallInput) (Skill, error) {
	if strings.TrimSpace(input.ProjectID) == "" || strings.TrimSpace(input.Imported.Name) == "" {
		return Skill{}, ErrInvalidInput
	}
	now := nowString()
	skillID := newID("skill")

	err := s.inTx(ctx, func(q *store.Queries) error {
		if _, err := q.GetStoryProjectByID(ctx, input.ProjectID); err != nil {
			return fmt.Errorf("get story project: %w", err)
		}
		if existing, err := q.GetSkillByProjectName(ctx, store.GetSkillByProjectNameParams{
			ProjectID: input.ProjectID,
			Name:      input.Imported.Name,
		}); err == nil {
			skillID = existing.ID
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("get existing skill: %w", err)
		}

		if err := q.UpsertSkill(ctx, store.UpsertSkillParams{
			ID:                   skillID,
			ProjectID:            input.ProjectID,
			Name:                 input.Imported.Name,
			DisplayName:          emptyDefault(input.Imported.DisplayName, input.Imported.Name),
			Description:          input.Imported.Description,
			InstructionsMarkdown: input.Imported.InstructionsMarkdown,
			SourceType:           string(input.SourceType),
			SourceLabel:          emptyDefault(input.SourceLabel, input.Imported.SourceLabel),
			ScriptCount:          input.Imported.ScriptCount,
			ScriptsDisabled:      boolInt(input.Imported.ScriptsDisabled),
			MetadataJson:         defaultJSON(input.Imported.MetadataJSON),
			InstalledAt:          now,
			UpdatedAt:            now,
		}); err != nil {
			return fmt.Errorf("upsert skill: %w", err)
		}
		if err := q.DeleteSkillFiles(ctx, skillID); err != nil {
			return fmt.Errorf("delete old skill files: %w", err)
		}
		if err := q.DeleteSkillRoutingHints(ctx, skillID); err != nil {
			return fmt.Errorf("delete old routing hints: %w", err)
		}
		for _, file := range input.Imported.Files {
			if err := q.CreateSkillFile(ctx, store.CreateSkillFileParams{
				ID:             newID("skillfile"),
				SkillID:        skillID,
				RelativePath:   file.RelativePath,
				Purpose:        string(file.Purpose),
				MediaType:      emptyDefault(file.MediaType, "text/plain; charset=utf-8"),
				BodyText:       file.BodyText,
				Bytes:          file.Bytes,
				ScriptDisabled: boolInt(file.ScriptDisabled),
				CreatedAt:      now,
			}); err != nil {
				return fmt.Errorf("create skill file %s: %w", file.RelativePath, err)
			}
		}
		for _, hint := range input.Imported.RoutingHints {
			if err := q.CreateSkillRoutingHint(ctx, store.CreateSkillRoutingHintParams{
				ID:          newID("skillroute"),
				SkillID:     skillID,
				ActionKind:  hint.ActionKind,
				ContentKind: hint.ContentKind,
				Tag:         hint.Tag,
				Priority:    hint.Priority,
				CreatedAt:   now,
			}); err != nil {
				return fmt.Errorf("create routing hint: %w", err)
			}
		}
		metadataJSON, err := marshalJSON(map[string]any{
			"skillId":         skillID,
			"name":            input.Imported.Name,
			"scriptCount":     input.Imported.ScriptCount,
			"scriptsDisabled": input.Imported.ScriptsDisabled,
			"sourceType":      string(input.SourceType),
		})
		if err != nil {
			return err
		}
		if err := q.CreateActivityEvent(ctx, store.CreateActivityEventParams{
			ID:           newID("event"),
			ProjectID:    input.ProjectID,
			SessionID:    sql.NullString{},
			EventType:    "skill_imported",
			Summary:      "Imported skill " + input.Imported.Name,
			MetadataJson: metadataJSON,
			CreatedAt:    now,
		}); err != nil {
			return fmt.Errorf("create skill imported event: %w", err)
		}
		return nil
	})
	if err != nil {
		return Skill{}, err
	}
	return s.Get(ctx, input.ProjectID, skillID)
}

func (s *Service) List(ctx context.Context, projectID string) ([]Skill, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, ErrInvalidInput
	}
	rows, err := s.queries.ListSkillsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	return s.skillsFromStore(ctx, s.queries, rows, false)
}

func (s *Service) Get(ctx context.Context, projectID, skillID string) (Skill, error) {
	if strings.TrimSpace(projectID) == "" || strings.TrimSpace(skillID) == "" {
		return Skill{}, ErrInvalidInput
	}
	row, err := s.queries.GetSkillByProjectID(ctx, store.GetSkillByProjectIDParams{
		ProjectID: projectID,
		ID:        skillID,
	})
	if err != nil {
		return Skill{}, fmt.Errorf("get skill: %w", err)
	}
	return s.skillFromStore(ctx, s.queries, row, true)
}

func (s *Service) GetByName(ctx context.Context, projectID, name string) (Skill, error) {
	if strings.TrimSpace(projectID) == "" || strings.TrimSpace(name) == "" {
		return Skill{}, ErrInvalidInput
	}
	row, err := s.queries.GetSkillByProjectName(ctx, store.GetSkillByProjectNameParams{
		ProjectID: projectID,
		Name:      name,
	})
	if err != nil {
		return Skill{}, fmt.Errorf("get skill by name: %w", err)
	}
	return s.skillFromStore(ctx, s.queries, row, true)
}

func (s *Service) ListRoutable(ctx context.Context, projectID, actionKind, contentKind string) ([]Skill, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, ErrInvalidInput
	}
	rows, err := s.queries.ListRoutableSkills(ctx, store.ListRoutableSkillsParams{
		ProjectID:   projectID,
		ActionKind:  actionKind,
		ContentKind: contentKind,
	})
	if err != nil {
		return nil, fmt.Errorf("list routable skills: %w", err)
	}
	return s.skillsFromStore(ctx, s.queries, rows, false)
}

func (s *Service) SelectSessionSkills(ctx context.Context, input SelectSessionSkillsInput) ([]Skill, error) {
	if strings.TrimSpace(input.ProjectID) == "" || strings.TrimSpace(input.SessionID) == "" {
		return nil, ErrInvalidInput
	}
	now := nowString()

	err := s.inTx(ctx, func(q *store.Queries) error {
		if _, err := q.GetAgentSession(ctx, store.GetAgentSessionParams{
			ProjectID: input.ProjectID,
			ID:        input.SessionID,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrInvalidInput
			}
			return fmt.Errorf("get agent session: %w", err)
		}
		for _, skillID := range input.SkillIDs {
			if strings.TrimSpace(skillID) == "" {
				return ErrInvalidInput
			}
			if _, err := q.GetSkillByProjectID(ctx, store.GetSkillByProjectIDParams{
				ProjectID: input.ProjectID,
				ID:        skillID,
			}); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return ErrInvalidInput
				}
				return fmt.Errorf("get skill: %w", err)
			}
		}
		if err := q.ReplaceSessionSkillsDelete(ctx, input.SessionID); err != nil {
			return fmt.Errorf("replace session skills: %w", err)
		}
		seen := make(map[string]struct{}, len(input.SkillIDs))
		for _, skillID := range input.SkillIDs {
			if _, ok := seen[skillID]; ok {
				continue
			}
			seen[skillID] = struct{}{}
			rowsAffected, err := q.AddSessionSkill(ctx, store.AddSessionSkillParams{
				ProjectID:  input.ProjectID,
				SessionID:  input.SessionID,
				SkillID:    skillID,
				SelectedAt: now,
			})
			if err != nil {
				return fmt.Errorf("add session skill: %w", err)
			}
			if rowsAffected == 0 {
				return ErrInvalidInput
			}
			selected, err := q.GetSkillByProjectID(ctx, store.GetSkillByProjectIDParams{
				ProjectID: input.ProjectID,
				ID:        skillID,
			})
			if err != nil {
				return fmt.Errorf("get selected skill: %w", err)
			}
			metadataJSON, err := marshalJSON(map[string]any{
				"skillId":         selected.ID,
				"name":            selected.Name,
				"scriptCount":     selected.ScriptCount,
				"scriptsDisabled": selected.ScriptsDisabled != 0,
			})
			if err != nil {
				return err
			}
			if err := q.CreateActivityEvent(ctx, store.CreateActivityEventParams{
				ID:           newID("event"),
				ProjectID:    input.ProjectID,
				SessionID:    sql.NullString{String: input.SessionID, Valid: true},
				EventType:    "skill_selected",
				Summary:      "Selected skill " + selected.Name,
				MetadataJson: metadataJSON,
				CreatedAt:    now,
			}); err != nil {
				return fmt.Errorf("create skill selected event: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.ListSessionSkills(ctx, input.ProjectID, input.SessionID)
}

func (s *Service) ListSessionSkills(ctx context.Context, projectID, sessionID string) ([]Skill, error) {
	if strings.TrimSpace(projectID) == "" || strings.TrimSpace(sessionID) == "" {
		return nil, ErrInvalidInput
	}
	rows, err := s.queries.ListSessionSkills(ctx, store.ListSessionSkillsParams{
		SessionID: sessionID,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("list session skills: %w", err)
	}
	return s.skillsFromStore(ctx, s.queries, rows, false)
}

func (s *Service) RenderForModel(ctx context.Context, input RenderSkillInput) (string, Skill, error) {
	skill, err := s.Get(ctx, input.ProjectID, input.SkillID)
	if err != nil {
		return "", Skill{}, err
	}

	var b strings.Builder
	b.WriteString(`<skill_content name="`)
	b.WriteString(escape(skill.Name))
	b.WriteString(`">`)
	b.WriteString("\n# Skill: ")
	b.WriteString(escape(skill.Name))
	b.WriteString("\n\n<instructions>\n")
	b.WriteString(escape(skill.InstructionsMarkdown))
	b.WriteString("\n</instructions>\n\n")
	fmt.Fprintf(&b, `<script_status disabled="%t" count="%d">Script files are imported for reference only. Writer does not execute bundled skill scripts in v1.</script_status>`, skill.ScriptsDisabled, skill.ScriptCount)
	b.WriteString("\n\n<skill_files>\n")

	remaining := 40 * 1024
	truncated := false
	for _, file := range skill.Files {
		b.WriteString(`<file path="`)
		b.WriteString(escape(file.RelativePath))
		b.WriteString(`" purpose="`)
		b.WriteString(escape(string(file.Purpose)))
		b.WriteString(`"`)
		if file.ScriptDisabled {
			b.WriteString(` disabled="true"`)
		}
		b.WriteString(">")

		if file.ScriptDisabled || file.Purpose == FilePurposeScript {
			b.WriteString("Script file is disabled and not executable.")
		} else if remaining > 0 {
			body := file.BodyText
			if len(body) > remaining {
				body = body[:remaining]
				truncated = true
			}
			remaining -= len(body)
			b.WriteString(escape(body))
		} else {
			truncated = true
		}

		b.WriteString("</file>\n")
	}
	b.WriteString("</skill_files>\n")
	if truncated {
		b.WriteString("\n<truncated>Additional skill file content is stored in Writer but omitted from model-visible output.</truncated>\n")
	}
	b.WriteString("</skill_content>")

	return b.String(), skill, nil
}

func (s *Service) inTx(ctx context.Context, fn func(*store.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	committed = true

	return nil
}

func (s *Service) skillsFromStore(ctx context.Context, q *store.Queries, rows []store.Skill, includeBodies bool) ([]Skill, error) {
	result := make([]Skill, 0, len(rows))
	for _, row := range rows {
		skill, err := s.skillFromStore(ctx, q, row, includeBodies)
		if err != nil {
			return nil, err
		}
		result = append(result, skill)
	}
	return result, nil
}

func (s *Service) skillFromStore(ctx context.Context, q *store.Queries, row store.Skill, includeBodies bool) (Skill, error) {
	files, err := q.ListSkillFiles(ctx, row.ID)
	if err != nil {
		return Skill{}, fmt.Errorf("list skill files: %w", err)
	}
	hints, err := q.ListSkillRoutingHints(ctx, row.ID)
	if err != nil {
		return Skill{}, fmt.Errorf("list skill routing hints: %w", err)
	}

	skill := skillFromStore(row)
	if !includeBodies {
		skill.InstructionsMarkdown = ""
	}
	skill.Files = make([]SkillFile, 0, len(files))
	for _, file := range files {
		converted := skillFileFromStore(file)
		if !includeBodies {
			converted.BodyText = ""
		}
		skill.Files = append(skill.Files, converted)
	}
	skill.RoutingHints = make([]RoutingHint, 0, len(hints))
	for _, hint := range hints {
		skill.RoutingHints = append(skill.RoutingHints, routingHintFromStore(hint))
	}
	return skill, nil
}

func skillFromStore(row store.Skill) Skill {
	return Skill{
		ID:                   row.ID,
		ProjectID:            row.ProjectID,
		Name:                 row.Name,
		DisplayName:          row.DisplayName,
		Description:          row.Description,
		InstructionsMarkdown: row.InstructionsMarkdown,
		SourceType:           SourceType(row.SourceType),
		SourceLabel:          row.SourceLabel,
		ScriptCount:          row.ScriptCount,
		ScriptsDisabled:      row.ScriptsDisabled != 0,
		MetadataJSON:         row.MetadataJson,
		InstalledAt:          row.InstalledAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func skillFileFromStore(row store.SkillFile) SkillFile {
	return SkillFile{
		ID:             row.ID,
		SkillID:        row.SkillID,
		RelativePath:   row.RelativePath,
		Purpose:        FilePurpose(row.Purpose),
		MediaType:      row.MediaType,
		BodyText:       row.BodyText,
		Bytes:          row.Bytes,
		ScriptDisabled: row.ScriptDisabled != 0,
		CreatedAt:      row.CreatedAt,
	}
}

func routingHintFromStore(row store.SkillRoutingHint) RoutingHint {
	return RoutingHint{
		ID:          row.ID,
		SkillID:     row.SkillID,
		ActionKind:  row.ActionKind,
		ContentKind: row.ContentKind,
		Tag:         row.Tag,
		Priority:    row.Priority,
		CreatedAt:   row.CreatedAt,
	}
}

func boolInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func defaultJSON(value string) string {
	if value == "" {
		return "{}"
	}
	return value
}

func marshalJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}
	return string(data), nil
}

func emptyDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func escape(value string) string {
	return html.EscapeString(value)
}

var idCounter uint64

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UTC().UnixNano(), atomic.AddUint64(&idCounter, 1))
}

func nowString() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
