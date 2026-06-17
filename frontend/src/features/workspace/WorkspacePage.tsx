import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { shallowEqual, useDispatch, useSelector } from "react-redux";
import { useNavigate, useParams } from "react-router-dom";

import { createContent, listContent, listProjects } from "../../api";
import { Button } from "../../shared/ui/button";
import type { ContentItem, ContentKind, StoryProject } from "../../types";
import { assistantActions } from "../assistant/assistantSlice";
import { modelSettingsActions } from "../model-settings/modelSettingsSlice";
import { reviewActions } from "../review/reviewSlice";
import { scriptRuntimeActions } from "../script-runtime/scriptRuntimeSlice";
import { skillsActions } from "../skills/skillsSlice";
import {
  loadProjectWorkspaceState,
  saveProjectWorkspaceState,
} from "./workspacePersistence";
import { selectWorkspaceProjectState } from "./workspaceSelectors";
import { workspaceActions, type WorkspaceState } from "./workspaceSlice";
import { WorkspaceShell } from "./WorkspaceShell";

type WorkspaceRouteParams = {
  projectId?: string;
  contentKind?: string;
  contentId?: string;
};

type WorkspaceRouteIdentity = WorkspaceRouteParams;

type WorkspaceRootState = {
  workspace: WorkspaceState;
};

const contentKinds = new Set<ContentKind>([
  "chapter",
  "story_bible_entry",
  "writing_brief",
  "project_note",
]);

function toContentKind(value: string): ContentKind {
  return contentKinds.has(value as ContentKind) ? (value as ContentKind) : "chapter";
}

function isSameRouteIdentity(left: WorkspaceRouteIdentity, right: WorkspaceRouteIdentity): boolean {
  return (
    left.projectId === right.projectId &&
    left.contentKind === right.contentKind &&
    left.contentId === right.contentId
  );
}

export function WorkspacePage() {
  const { projectId, contentKind: routeContentKind, contentId } = useParams<WorkspaceRouteParams>();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const workspace = useSelector((state: WorkspaceRootState) => state.workspace);
  const projectWorkspaceState = useSelector(selectWorkspaceProjectState, shallowEqual);
  const [hydratedProjectId, setHydratedProjectId] = useState<string | null>(null);
  const [projects, setProjects] = useState<StoryProject[]>([]);
  const [projectError, setProjectError] = useState<string | null>(null);
  const [contentItems, setContentItems] = useState<ContentItem[]>([]);
  const [contentError, setContentError] = useState<string | null>(null);
  const [contentCreateError, setContentCreateError] = useState<string | null>(null);
  const [projectsLoading, setProjectsLoading] = useState(true);
  const [contentLoading, setContentLoading] = useState(false);
  const [contentCreating, setContentCreating] = useState(false);
  const createAbortControllerRef = useRef<AbortController | null>(null);
  const routeIdentityRef = useRef<WorkspaceRouteIdentity>({ projectId, contentKind: routeContentKind, contentId });
  routeIdentityRef.current = { projectId, contentKind: routeContentKind, contentId };
  const contentKind = routeContentKind
    ? toContentKind(routeContentKind)
    : toContentKind(workspace.lastContentKind);

  useEffect(() => {
    let ignore = false;
    setProjectsLoading(true);
    setProjectError(null);

    void listProjects()
      .then((items) => {
        if (!ignore) setProjects(items);
      })
      .catch((cause: unknown) => {
        if (!ignore) {
          setProjectError(cause instanceof Error ? cause.message : "Could not load projects");
        }
      })
      .finally(() => {
        if (!ignore) setProjectsLoading(false);
      });

    return () => {
      ignore = true;
    };
  }, []);

  useLayoutEffect(() => {
    if (!projectId) return;
    dispatch(assistantActions.resetForProject());
    dispatch(modelSettingsActions.resetForProject());
    dispatch(reviewActions.resetForProject());
    dispatch(scriptRuntimeActions.resetForProject());
    dispatch(skillsActions.resetForProject());
    dispatch(workspaceActions.hydrateProjectState(loadProjectWorkspaceState(projectId)));
    setHydratedProjectId(projectId);
  }, [dispatch, projectId]);

  useEffect(() => {
    if (!routeContentKind) return;
    dispatch(workspaceActions.setLastContentKind(toContentKind(routeContentKind)));
  }, [dispatch, routeContentKind]);

  useEffect(() => {
    abortPendingCreate();
  }, [routeContentKind, contentId]);

  useEffect(() => {
    setContentCreating(false);

    return () => {
      createAbortControllerRef.current?.abort();
      createAbortControllerRef.current = null;
    };
  }, [projectId]);

  useEffect(() => {
    if (!projectId || hydratedProjectId !== projectId) return;
    saveProjectWorkspaceState(projectId, projectWorkspaceState);
  }, [hydratedProjectId, projectId, projectWorkspaceState]);

  useEffect(() => {
    if (!projectId) {
      setContentItems([]);
      return;
    }

    const abortController = new AbortController();
    setContentLoading(true);
    setContentError(null);
    setContentCreateError(null);
    setContentItems([]);

    void listContent(projectId, contentKind, abortController.signal)
      .then((items) => {
        if (abortController.signal.aborted) return;
        setContentItems(items);
      })
      .catch((cause: unknown) => {
        if (abortController.signal.aborted) return;
        setContentError(cause instanceof Error ? cause.message : "Could not load content");
      })
      .finally(() => {
        if (!abortController.signal.aborted) setContentLoading(false);
      });

    return () => abortController.abort();
  }, [contentKind, projectId]);

  const activeProject = useMemo(
    () => projects.find((project) => project.id === projectId) ?? null,
    [projectId, projects],
  );

  const selectedContent = useMemo(() => {
    if (contentItems.length === 0) return null;

    if (contentId) {
      return contentItems.find((item) => item.id === contentId) ?? null;
    }

    if (workspace.fallbackContentId) {
      const fallback = contentItems.find((item) => item.id === workspace.fallbackContentId);
      if (fallback) return fallback;
    }

    return contentItems[0] ?? null;
  }, [contentId, contentItems, workspace.fallbackContentId]);

  function contentKindLabel(kind: ContentKind): string {
    switch (kind) {
      case "chapter":
        return "Chapter";
      case "story_bible_entry":
        return "Story bible entry";
      case "writing_brief":
        return "Writing brief";
      case "project_note":
        return "Project note";
    }
  }

  function defaultBodyForKind(kind: ContentKind): string {
    switch (kind) {
      case "chapter":
        return "";
      case "story_bible_entry":
        return "## Overview\n\n";
      case "writing_brief":
        return "## Brief\n\n";
      case "project_note":
        return "## Note\n\n";
    }
  }

  function abortPendingCreate(): void {
    const abortController = createAbortControllerRef.current;
    if (!abortController) return;
    abortController.abort();
    createAbortControllerRef.current = null;
    setContentCreating(false);
  }

  function isCurrentCreateContext(
    abortController: AbortController,
    routeIdentity: WorkspaceRouteIdentity,
  ): boolean {
    return (
      !abortController.signal.aborted &&
      createAbortControllerRef.current === abortController &&
      isSameRouteIdentity(routeIdentityRef.current, routeIdentity)
    );
  }

  function handleCreateContent(kind: ContentKind): void {
    if (!projectId || contentLoading || createAbortControllerRef.current) return;
    const abortController = new AbortController();
    const routeIdentity = routeIdentityRef.current;
    createAbortControllerRef.current = abortController;
    setContentCreating(true);
    setContentCreateError(null);

    void (async () => {
      const targetItems = contentKind === kind
        ? contentItems
        : await listContent(projectId, kind, abortController.signal);
      const nextNumber = targetItems.length + 1;

      return createContent(projectId, {
        kind,
        title: `${contentKindLabel(kind)} ${nextNumber}`,
        bodyMarkdown: defaultBodyForKind(kind),
        metadataJson: "{}",
        sortOrder: nextNumber,
        reason: "created from workspace",
      }, abortController.signal);
    })()
      .then((item) => {
        if (!isCurrentCreateContext(abortController, routeIdentity)) return;
        createAbortControllerRef.current = null;
        setContentCreating(false);
        setContentItems((items) => (item.kind === contentKind ? [...items, item] : items));
        dispatch(workspaceActions.setLastContentKind(item.kind));
        dispatch(workspaceActions.setFallbackContentId(item.id));
        navigate(
          `/projects/${encodeURIComponent(projectId)}/content/${encodeURIComponent(item.kind)}/${encodeURIComponent(item.id)}`,
        );
      })
      .catch((cause: unknown) => {
        if (!isCurrentCreateContext(abortController, routeIdentity)) return;
        setContentCreateError(cause instanceof Error ? cause.message : "Could not create content");
      })
      .finally(() => {
        if (!isCurrentCreateContext(abortController, routeIdentity)) return;
        createAbortControllerRef.current = null;
        setContentCreating(false);
      });
  }

  function handleSelectContent(item: ContentItem): void {
    if (!projectId) return;
    abortPendingCreate();
    dispatch(workspaceActions.setFallbackContentId(item.id));
    navigate(
      `/projects/${encodeURIComponent(projectId)}/content/${encodeURIComponent(item.kind)}/${encodeURIComponent(item.id)}`,
    );
  }

  function handleContentKindChange(kind: ContentKind): void {
    if (!projectId) return;
    abortPendingCreate();
    dispatch(workspaceActions.setLastContentKind(kind));
    navigate(`/projects/${encodeURIComponent(projectId)}`);
  }

  function handleContentSaved(item: ContentItem): void {
    setContentItems((items) =>
      items.map((existing) => (existing.id === item.id ? item : existing)),
    );
  }

  if (!projectId) {
    return (
      <main className="flex min-h-dvh items-center justify-center bg-background p-6 text-foreground">
        <p className="text-sm text-muted-foreground">Project route is missing.</p>
      </main>
    );
  }

  if (projectsLoading) {
    return (
      <main className="flex min-h-dvh items-center justify-center bg-background p-6 text-foreground">
        <p className="text-sm text-muted-foreground">Loading workspace...</p>
      </main>
    );
  }

  if (projectError) {
    return (
      <main className="flex min-h-dvh items-center justify-center bg-background p-6 text-foreground">
        <div className="flex flex-col gap-3 rounded-md border border-border bg-background p-4">
          <p role="alert" className="text-sm text-destructive">
            Could not load workspace: {projectError}
          </p>
          <Button type="button" variant="outline" onClick={() => navigate("/projects")}>
            Back to projects
          </Button>
        </div>
      </main>
    );
  }

  return (
    <WorkspaceShell
      projectId={projectId}
      projectTitle={activeProject?.title ?? "Untitled project"}
      contentItems={contentItems}
      contentLoading={contentLoading}
      contentError={contentError}
      contentCreateError={contentCreateError}
      contentCreating={contentCreating}
      activeContentKind={contentKind}
      selectedContent={selectedContent}
      onCreateContent={handleCreateContent}
      onSelectContent={handleSelectContent}
      onContentKindChange={handleContentKindChange}
      onContentSaved={handleContentSaved}
    />
  );
}
