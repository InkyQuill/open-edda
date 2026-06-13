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
