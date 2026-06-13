export type SkillFilePurpose = "instruction" | "template" | "reference" | "data" | "script" | "other";

export type SkillFile = {
  id: string;
  skillId: string;
  relativePath: string;
  purpose: SkillFilePurpose;
  mediaType: string;
  bodyText?: string;
  bytes: number;
  scriptDisabled: boolean;
  createdAt: string;
};

export type RoutingHint = {
  id?: string;
  skillId?: string;
  actionKind: string;
  contentKind: string;
  tag: string;
  priority: number;
  createdAt?: string;
};

export type WriterSkill = {
  id: string;
  projectId: string;
  name: string;
  displayName: string;
  description: string;
  instructionsMarkdown?: string;
  sourceType: "upload" | "local_directory";
  sourceLabel: string;
  scriptCount: number;
  scriptsDisabled: boolean;
  metadataJson: string;
  installedAt: string;
  updatedAt: string;
  files?: SkillFile[];
  routingHints?: RoutingHint[];
};
