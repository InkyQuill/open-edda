package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/skill"
)

const fictionAssistantSystemPrompt = "You are a fiction writing assistant working inside Edda. Preserve the author's intent, respect established project facts, and use available tools to inspect project context before making claims. Do not invent durable worldbuilding facts unless the author asks you to brainstorm."

func (s *Service) BuildActionPrompt(ctx context.Context, input BuildPromptInput) (PromptBundle, error) {
	if err := validateActionKind(input.ActionKind); err != nil {
		return PromptBundle{}, err
	}
	if _, err := s.queries.GetStoryProjectByID(ctx, input.ProjectID); err != nil {
		return PromptBundle{}, fmt.Errorf("get story project: %w", err)
	}
	if input.SessionID != "" {
		if _, err := s.GetSession(ctx, input.ProjectID, input.SessionID); err != nil {
			return PromptBundle{}, fmt.Errorf("validate session: %w", err)
		}
	}
	if input.TargetContentID == "" {
		return PromptBundle{}, fmt.Errorf("target content ID is required")
	}

	profile, err := s.GetPromptProfile(ctx, input.ProjectID)
	if err != nil {
		return PromptBundle{}, err
	}

	target, err := s.projectService.GetContent(ctx, input.ProjectID, input.TargetContentID)
	if err != nil {
		return PromptBundle{}, err
	}
	if target.Kind != project.KindChapter {
		return PromptBundle{}, fmt.Errorf("target content must be chapter, got %q", target.Kind)
	}

	writingBriefs, err := s.layeredWritingBriefs(ctx, input.ProjectID, input.TargetContentID)
	if err != nil {
		return PromptBundle{}, err
	}

	tools := QuickActionToolDefinitions()
	sources, err := buildContextSources(profile, writingBriefs, target, input, tools)
	if err != nil {
		return PromptBundle{}, err
	}
	if s.skillService != nil {
		skillSources, err := s.buildSkillContextSources(ctx, input.ProjectID, input.SessionID)
		if err != nil {
			return PromptBundle{}, err
		}
		sources = append(sources, skillSources...)
	}

	developerMessage := renderDeveloperMessage(sources)
	userMessage := renderUserMessage(input)
	metadataJSON, err := marshalPromptMetadata(input, sources)
	if err != nil {
		return PromptBundle{}, err
	}

	return PromptBundle{
		SystemMessage:    fictionAssistantSystemPrompt,
		DeveloperMessage: developerMessage,
		UserMessage:      userMessage,
		Tools:            tools,
		MetadataJSON:     metadataJSON,
		ContextSources:   sources,
	}, nil
}

func (s *Service) buildSkillContextSources(ctx context.Context, projectID, sessionID string) ([]ContextSourceSnapshot, error) {
	installed, err := s.skillService.List(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	available := describedSkillSummaries(installed)
	availableRendered := renderAvailableSkills(available, len(installed))
	availableValue, err := marshalJSON(map[string]any{
		"skills": available,
	})
	if err != nil {
		return nil, err
	}

	sources := []ContextSourceSnapshot{
		{
			SourceKey:        "available_skills",
			SourceVersion:    skillsVersion(installed),
			RenderedMarkdown: availableRendered,
			ValueJSON:        availableValue,
		},
	}
	if sessionID == "" {
		return sources, nil
	}

	selectedSkills, err := s.skillService.ListSessionSkills(ctx, projectID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list session skills: %w", err)
	}
	if len(selectedSkills) == 0 {
		return sources, nil
	}
	selected, err := s.selectedSkillPromptSummaries(ctx, projectID, selectedSkills)
	if err != nil {
		return nil, err
	}
	selectedRendered := renderSelectedSkills(selected)
	selectedValue, err := marshalJSON(map[string]any{
		"skills": selected,
	})
	if err != nil {
		return nil, err
	}
	sources = append(sources, ContextSourceSnapshot{
		SourceKey:        "selected_skills",
		SourceVersion:    skillsVersion(selectedSkills),
		RenderedMarkdown: selectedRendered,
		ValueJSON:        selectedValue,
	})
	return sources, nil
}

func (s *Service) layeredWritingBriefs(ctx context.Context, projectID, targetContentID string) ([]project.ContentItem, error) {
	items, err := s.projectService.ListContent(ctx, projectID, project.KindWritingBrief)
	if err != nil {
		return nil, err
	}

	projectBriefs := make([]project.ContentItem, 0)
	targetBriefs := make([]project.ContentItem, 0)
	for _, item := range items {
		metadata, err := writingBriefMetadata(item.MetadataJSON)
		if err != nil {
			return nil, fmt.Errorf("parse writing brief metadata for %q: %w", item.ID, err)
		}
		switch {
		case metadata.ContentItemID == "":
			projectBriefs = append(projectBriefs, item)
		case metadata.ContentItemID == targetContentID:
			targetBriefs = append(targetBriefs, item)
		}
	}
	sortWritingBriefsForPrompt(projectBriefs)
	sortWritingBriefsForPrompt(targetBriefs)

	result := make([]project.ContentItem, 0, len(projectBriefs)+len(targetBriefs))
	result = append(result, projectBriefs...)
	result = append(result, targetBriefs...)
	return result, nil
}

type briefMetadata struct {
	ContentItemID string `json:"contentItemId"`
	Scope         string `json:"scope"`
}

func writingBriefMetadata(value string) (briefMetadata, error) {
	if strings.TrimSpace(value) == "" {
		return briefMetadata{}, nil
	}
	var metadata briefMetadata
	if err := json.Unmarshal([]byte(value), &metadata); err != nil {
		return briefMetadata{}, err
	}
	switch metadata.Scope {
	case "":
		return metadata, nil
	case "project":
		if metadata.ContentItemID != "" {
			return briefMetadata{}, fmt.Errorf("project-scoped writing brief must not have contentItemId")
		}
	case "chapter":
		if metadata.ContentItemID == "" {
			return briefMetadata{}, fmt.Errorf("chapter-scoped writing brief must have contentItemId")
		}
	default:
		return briefMetadata{}, fmt.Errorf("unknown writing brief scope %q", metadata.Scope)
	}
	return metadata, nil
}

func sortWritingBriefsForPrompt(briefs []project.ContentItem) {
	sort.Slice(briefs, func(i, j int) bool {
		if briefs[i].SortOrder != briefs[j].SortOrder {
			return briefs[i].SortOrder < briefs[j].SortOrder
		}
		if briefs[i].Title != briefs[j].Title {
			return briefs[i].Title < briefs[j].Title
		}
		return briefs[i].ID < briefs[j].ID
	})
}

func buildContextSources(profile PromptProfile, writingBriefs []project.ContentItem, target project.ContentItem, input BuildPromptInput, tools []CompletionTool) ([]ContextSourceSnapshot, error) {
	profileRendered := renderPromptProfile(profile)
	profileValue, err := marshalJSON(map[string]any{
		"id":                        profile.ID,
		"projectId":                 profile.ProjectID,
		"genre":                     profile.Genre,
		"tense":                     profile.Tense,
		"pov":                       profile.POV,
		"voice":                     profile.Voice,
		"instructionsMarkdown":      profile.InstructionsMarkdown,
		"promptRecordRetentionDays": profile.PromptRecordRetentionDays,
	})
	if err != nil {
		return nil, err
	}

	briefsRendered := renderWritingBriefs(writingBriefs)
	briefsValue, err := marshalJSON(writingBriefValues(writingBriefs))
	if err != nil {
		return nil, err
	}

	targetRendered := renderTargetContentWindow(target, input.CursorSummary)
	targetValue, err := marshalJSON(map[string]any{
		"id":              target.ID,
		"kind":            target.Kind,
		"title":           target.Title,
		"currentRevision": target.CurrentRevision,
		"cursorSummary":   input.CursorSummary,
	})
	if err != nil {
		return nil, err
	}

	providerRendered := "## Provider Disclosure\n\nUse tools for additional project context instead of assuming the whole project is present in this prompt."
	providerValue, err := marshalJSON(map[string]any{
		"wholeProjectInPrompt": false,
		"instruction":          "Use tools for additional project context instead of assuming the whole project is present in this prompt.",
	})
	if err != nil {
		return nil, err
	}

	toolRendered := "## Tool Catalog\n\nProject context tools are available for this action. Use them to inspect worldbuilding, other chapters, notes, skills, and revisions when needed before generating the final action output. Direct write tools are not available in this action request; the action pipeline handles preview or direct apply after the final response."
	toolValue, err := marshalJSON(map[string]any{
		"tools": tools,
	})
	if err != nil {
		return nil, err
	}

	return []ContextSourceSnapshot{
		{
			SourceKey:        "prompt_profile",
			SourceVersion:    profile.UpdatedAt,
			RenderedMarkdown: profileRendered,
			ValueJSON:        profileValue,
		},
		{
			SourceKey:        "writing_briefs",
			SourceVersion:    writingBriefsVersion(writingBriefs),
			RenderedMarkdown: briefsRendered,
			ValueJSON:        briefsValue,
		},
		{
			SourceKey:        "target_content_window",
			SourceVersion:    fmt.Sprintf("revision:%d", target.CurrentRevision),
			RenderedMarkdown: targetRendered,
			ValueJSON:        targetValue,
		},
		{
			SourceKey:        "provider_disclosure",
			RenderedMarkdown: providerRendered,
			ValueJSON:        providerValue,
		},
		{
			SourceKey:        "tool_catalog",
			RenderedMarkdown: toolRendered,
			ValueJSON:        toolValue,
		},
	}, nil
}

func renderPromptProfile(profile PromptProfile) string {
	var b strings.Builder
	b.WriteString("## Prompt Profile\n\n")
	b.WriteString("Genre: " + profile.Genre + "\n")
	b.WriteString("Tense: " + profile.Tense + "\n")
	b.WriteString("Point of view: " + profile.POV + "\n")
	b.WriteString("Voice: " + profile.Voice + "\n")
	if profile.InstructionsMarkdown != "" {
		b.WriteString("\n### Writing Instructions\n\n")
		b.WriteString(profile.InstructionsMarkdown)
		b.WriteString("\n")
	}
	return b.String()
}

func renderWritingBriefs(briefs []project.ContentItem) string {
	var b strings.Builder
	b.WriteString("## Writing Briefs\n")
	for _, brief := range briefs {
		b.WriteString("\n### ")
		b.WriteString(brief.Title)
		b.WriteString("\n\n")
		b.WriteString(brief.BodyMarkdown)
		b.WriteString("\n")
	}
	if len(briefs) == 0 {
		b.WriteString("\nNo writing briefs are available for this target.\n")
	}
	return b.String()
}

func renderTargetContentWindow(target project.ContentItem, cursorSummary string) string {
	var b strings.Builder
	b.WriteString("## Target Content Window\n\n")
	b.WriteString("Target chapter: ")
	b.WriteString(target.Title)
	b.WriteString("\n")
	if cursorSummary != "" {
		b.WriteString("Cursor location: ")
		b.WriteString(cursorSummary)
		b.WriteString("\n")
	}
	if target.BodyMarkdown != "" {
		b.WriteString("\n### Current Text\n\n")
		b.WriteString(target.BodyMarkdown)
		b.WriteString("\n")
	}
	return b.String()
}

func describedSkillSummaries(skills []skill.Skill) []SkillSummary {
	summaries := make([]SkillSummary, 0, len(skills))
	for _, item := range skills {
		if strings.TrimSpace(item.Description) == "" {
			continue
		}
		summaries = append(summaries, skillSummary(item))
	}
	return summaries
}

func skillSummaries(skills []skill.Skill) []SkillSummary {
	summaries := make([]SkillSummary, 0, len(skills))
	for _, item := range skills {
		summaries = append(summaries, skillSummary(item))
	}
	return summaries
}

func skillSummary(item skill.Skill) SkillSummary {
	metadata := skillPromptMetadata(item.MetadataJSON)
	return SkillSummary{
		ID:              item.ID,
		Name:            item.Name,
		Description:     item.Description,
		UseCases:        metadata.UseCases,
		DoNotUse:        metadata.DoNotUse,
		ScriptCount:     item.ScriptCount,
		ScriptsDisabled: item.ScriptsDisabled,
	}
}

type skillPromptMetadataValue struct {
	UseCases []string `json:"useCases"`
	DoNotUse []string `json:"doNotUse"`
}

func skillPromptMetadata(value string) skillPromptMetadataValue {
	var metadata skillPromptMetadataValue
	if strings.TrimSpace(value) == "" {
		return metadata
	}
	_ = json.Unmarshal([]byte(value), &metadata)
	return metadata
}

type selectedSkillPromptSummary struct {
	SkillSummary
	EnabledScripts []selectedScriptPromptSummary `json:"enabledScripts,omitempty"`
}

type selectedScriptPromptSummary struct {
	Path    string `json:"path"`
	Runtime string `json:"runtime"`
}

func (s *Service) selectedSkillPromptSummaries(ctx context.Context, projectID string, skills []skill.Skill) ([]selectedSkillPromptSummary, error) {
	summaries := make([]selectedSkillPromptSummary, 0, len(skills))
	for _, item := range skills {
		summary := selectedSkillPromptSummary{SkillSummary: skillSummary(item)}
		audits, err := s.skillService.ListScriptAudits(ctx, projectID, item.ID)
		if err != nil {
			return nil, fmt.Errorf("list script audits: %w", err)
		}
		for _, audit := range audits {
			if audit.Approval == nil || !audit.Approval.Enabled {
				continue
			}
			summary.EnabledScripts = append(summary.EnabledScripts, selectedScriptPromptSummary{
				Path:    audit.RelativePath,
				Runtime: audit.Runtime,
			})
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func renderAvailableSkills(skills []SkillSummary, installedCount int) string {
	var b strings.Builder
	b.WriteString("## Available Skills\n\n")
	b.WriteString("Skills provide specialized instructions and workflows for specific writing tasks.\n")
	b.WriteString("Use the skill tool to load a skill when the task matches its description.\n")
	if installedCount == 0 {
		b.WriteString("\nNo skills are installed for this project.\n")
		return b.String()
	}
	if len(skills) == 0 {
		b.WriteString("\nNo described skills are available for this project.\n")
		return b.String()
	}
	b.WriteString("\n<available_skills>\n")
	for _, item := range skills {
		b.WriteString("  <skill>\n")
		b.WriteString("    <id>" + escapePromptXML(item.ID) + "</id>\n")
		b.WriteString("    <name>" + escapePromptXML(item.Name) + "</name>\n")
		b.WriteString("    <description>" + escapePromptXML(item.Description) + "</description>\n")
		renderSkillUseMetadata(&b, item.UseCases, item.DoNotUse, "    ")
		b.WriteString("  </skill>\n")
	}
	b.WriteString("</available_skills>\n")
	return b.String()
}

func renderSelectedSkills(skills []selectedSkillPromptSummary) string {
	var b strings.Builder
	b.WriteString("## Selected Skills\n\n")
	b.WriteString("The author selected these skills for this session.\n")
	b.WriteString("Enabled runtime helpers are available only through the `skill_script` tool. They return reports, proposals, generated data, or drafts and cannot directly apply project changes.\n")
	b.WriteString("Disabled scripts remain unavailable.\n")
	b.WriteString("<selected_skills>\n")
	for _, item := range skills {
		b.WriteString("  <skill>\n")
		b.WriteString("    <id>" + escapePromptXML(item.ID) + "</id>\n")
		b.WriteString("    <name>" + escapePromptXML(item.Name) + "</name>\n")
		if item.Description != "" {
			b.WriteString("    <description>" + escapePromptXML(item.Description) + "</description>\n")
		}
		renderSkillUseMetadata(&b, item.UseCases, item.DoNotUse, "    ")
		if len(item.EnabledScripts) == 0 {
			fmt.Fprintf(&b, "    <script_status disabled=\"%t\" count=\"%d\">Scripts are available only as inert reference files.</script_status>\n", item.ScriptsDisabled, item.ScriptCount)
		} else {
			fmt.Fprintf(&b, "    <script_status disabled=\"false\" count=\"%d\">Enabled runtime helpers must be invoked with skill_script.</script_status>\n", item.ScriptCount)
			for _, script := range item.EnabledScripts {
				b.WriteString("    <runtime_helper>\n")
				b.WriteString("      <script_path>" + escapePromptXML(script.Path) + "</script_path>\n")
				if script.Runtime != "" {
					b.WriteString("      <runtime>" + escapePromptXML(script.Runtime) + "</runtime>\n")
				}
				b.WriteString("    </runtime_helper>\n")
			}
		}
		b.WriteString("  </skill>\n")
	}
	b.WriteString("</selected_skills>\n")
	return b.String()
}

func renderSkillUseMetadata(b *strings.Builder, useCases []string, doNotUse []string, indent string) {
	if len(useCases) > 0 {
		b.WriteString(indent + "<use_cases>\n")
		for _, item := range useCases {
			if strings.TrimSpace(item) == "" {
				continue
			}
			b.WriteString(indent + "  <case>" + escapePromptXML(item) + "</case>\n")
		}
		b.WriteString(indent + "</use_cases>\n")
	}
	if len(doNotUse) > 0 {
		b.WriteString(indent + "<do_not_use>\n")
		for _, item := range doNotUse {
			if strings.TrimSpace(item) == "" {
				continue
			}
			b.WriteString(indent + "  <case>" + escapePromptXML(item) + "</case>\n")
		}
		b.WriteString(indent + "</do_not_use>\n")
	}
}

func skillsVersion(skills []skill.Skill) string {
	versions := make([]string, 0, len(skills))
	for _, item := range skills {
		versions = append(versions, item.ID+":"+item.UpdatedAt)
	}
	return strings.Join(versions, ",")
}

func escapePromptXML(value string) string {
	return html.EscapeString(value)
}

func renderDeveloperMessage(sources []ContextSourceSnapshot) string {
	var b strings.Builder
	for i, source := range sources {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(source.RenderedMarkdown)
	}
	return b.String()
}

func renderUserMessage(input BuildPromptInput) string {
	var b strings.Builder
	switch input.ActionKind {
	case ActionKindContinuation:
		b.WriteString("Continue the target chapter from the cursor location.")
	default:
		b.WriteString("Complete the requested writing action.")
	}
	if input.UserGuidance != "" {
		b.WriteString("\n\nUser guidance:\n")
		b.WriteString(input.UserGuidance)
	}
	return b.String()
}

func writingBriefValues(briefs []project.ContentItem) []map[string]any {
	values := make([]map[string]any, 0, len(briefs))
	for _, brief := range briefs {
		values = append(values, map[string]any{
			"id":              brief.ID,
			"title":           brief.Title,
			"metadataJson":    brief.MetadataJSON,
			"currentRevision": brief.CurrentRevision,
		})
	}
	return values
}

func writingBriefsVersion(briefs []project.ContentItem) string {
	versions := make([]string, 0, len(briefs))
	for _, brief := range briefs {
		versions = append(versions, fmt.Sprintf("%s:%d", brief.ID, brief.CurrentRevision))
	}
	return strings.Join(versions, ",")
}

func marshalPromptMetadata(input BuildPromptInput, sources []ContextSourceSnapshot) (string, error) {
	sourceKeys := make([]string, 0, len(sources))
	for _, source := range sources {
		sourceKeys = append(sourceKeys, source.SourceKey)
	}
	return marshalJSON(map[string]any{
		"projectId":       input.ProjectID,
		"sessionId":       input.SessionID,
		"actionKind":      input.ActionKind,
		"targetContentId": input.TargetContentID,
		"contextSources":  sourceKeys,
	})
}

func marshalJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal prompt JSON: %w", err)
	}
	return string(data), nil
}
