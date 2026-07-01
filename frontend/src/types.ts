export type StoryProject = {
  id: string;
  title: string;
  slug: string;
  language: string;
};

export type ContentKind = "chapter" | "story_bible_entry" | "writing_brief" | "project_note";

export type ContentItem = {
  id: string;
  projectId: string;
  kind: ContentKind;
  title: string;
  slug: string;
  bodyMarkdown: string;
  metadataJson: string;
  sortOrder: number;
  currentRevision: number;
};

export type Revision = {
  id: string;
  contentItemId: string;
  revisionNumber: number;
  bodyMarkdown: string;
  metadataJson: string;
  reason: string;
  createdBy: string;
  createdAt: string;
  agentSessionId?: string;
  actionKind?: string;
  modelVariantId?: string;
  skillId?: string;
};
