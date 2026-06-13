import { useEffect, useMemo, useState } from "react";
import type { ContentItem, ContentKind, StoryProject } from "./types";
import { listContent, listProjects } from "./api";
import "./styles.css";

const contentKinds: Array<{ kind: ContentKind; label: string }> = [
  { kind: "chapter", label: "Chapters" },
  { kind: "story_bible_entry", label: "Story Bible" },
  { kind: "writing_brief", label: "Briefs" },
  { kind: "project_note", label: "Notes" },
];

export function App() {
  const [projects, setProjects] = useState<StoryProject[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);
  const [selectedKind, setSelectedKind] = useState<ContentKind>("chapter");
  const [contentItems, setContentItems] = useState<ContentItem[]>([]);
  const [selectedContentId, setSelectedContentId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isContentLoading, setIsContentLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [contentError, setContentError] = useState<string | null>(null);

  useEffect(() => {
    void listProjects()
      .then((items) => {
        setProjects(items);
        setSelectedProjectId((current) => current ?? items[0]?.id ?? null);
        setError(null);
      })
      .catch((cause: unknown) => {
        setError(cause instanceof Error ? cause.message : "Project list failed");
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  useEffect(() => {
    if (!selectedProjectId) {
      setContentItems([]);
      setSelectedContentId(null);
      return;
    }

    const abortController = new AbortController();
    setIsContentLoading(true);
    setContentItems([]);
    setSelectedContentId(null);
    setContentError(null);
    void listContent(selectedProjectId, selectedKind, abortController.signal)
      .then((items) => {
        setContentItems(items);
        setSelectedContentId(items[0]?.id ?? null);
        setContentError(null);
      })
      .catch((cause: unknown) => {
        if (cause instanceof DOMException && cause.name === "AbortError") {
          return;
        }
        setContentItems([]);
        setSelectedContentId(null);
        setContentError(cause instanceof Error ? cause.message : "Content list failed");
      })
      .finally(() => {
        if (!abortController.signal.aborted) {
          setIsContentLoading(false);
        }
      });

    return () => {
      abortController.abort();
    };
  }, [selectedProjectId, selectedKind]);

  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) ?? null,
    [projects, selectedProjectId],
  );
  const selectedContent = useMemo(
    () => contentItems.find((item) => item.id === selectedContentId) ?? null,
    [contentItems, selectedContentId],
  );

  return (
    <main className="app-shell">
      <section className="project-dashboard">
        <header>
          <h1>Writer</h1>
          <p>Private AI writing studio</p>
        </header>

        <div className="project-list">
          {isLoading ? (
            <p>Loading story projects...</p>
          ) : error ? (
            <p role="alert">Could not load story projects.</p>
          ) : projects.length === 0 ? (
            <p>No story projects yet.</p>
          ) : (
            projects.map((project) => (
              <button
                key={project.id}
                className="project-row"
                type="button"
                data-active={project.id === selectedProjectId}
                aria-pressed={project.id === selectedProjectId}
                onClick={() => setSelectedProjectId(project.id)}
              >
                <span className="project-row-title">{project.title}</span>
                <span className="project-row-meta">{project.language || "Language not set"}</span>
              </button>
            ))
          )}
        </div>
      </section>

      {selectedProject ? (
        <section className="workspace-shell" aria-label={`${selectedProject.title} workspace`}>
          <aside className="workspace-nav" aria-label="Content type">
            {contentKinds.map((entry) => (
              <button
                key={entry.kind}
                type="button"
                data-active={entry.kind === selectedKind}
                aria-pressed={entry.kind === selectedKind}
                onClick={() => setSelectedKind(entry.kind)}
              >
                {entry.label}
              </button>
            ))}
          </aside>

          <section className="content-list" aria-label="Project content">
            <header>
              <h2>{contentKinds.find((entry) => entry.kind === selectedKind)?.label}</h2>
              <p>{selectedProject.title}</p>
            </header>
            {isContentLoading ? (
              <p>Loading content...</p>
            ) : contentError ? (
              <p role="alert">Could not load content.</p>
            ) : contentItems.length === 0 ? (
              <p>No content in this section yet.</p>
            ) : (
              contentItems.map((item) => (
                <button
                  key={item.id}
                  className="content-row"
                  type="button"
                  data-active={item.id === selectedContentId}
                  aria-pressed={item.id === selectedContentId}
                  onClick={() => setSelectedContentId(item.id)}
                >
                  <span>{item.title}</span>
                  <small>Revision {item.currentRevision}</small>
                </button>
              ))
            )}
          </section>

          <section className="detail-panel" aria-label="Content detail">
            {selectedContent ? (
              <>
                <header>
                  <h2>{selectedContent.title}</h2>
                  <p>{selectedContent.kind.replaceAll("_", " ")}</p>
                </header>
                <textarea readOnly value={selectedContent.bodyMarkdown} aria-label={`${selectedContent.title} markdown`} />
              </>
            ) : (
              <p>Select an item to preview its Markdown.</p>
            )}
          </section>
        </section>
      ) : null}
    </main>
  );
}
