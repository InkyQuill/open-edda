import type React from "react";
import { useEffect, useRef, useState } from "react";
import { ArrowUpRight, LogOut, Plus, Settings2, Upload } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";

import { createProject, importElysiumProject } from "../../api";
import { clearToken } from "../../authApi";
import { Button } from "../../shared/ui/button";
import { useProjects } from "./projectHooks";

export function ProjectsPage() {
  const navigate = useNavigate();
  const { projects, loading, error, reload } = useProjects();
  const importInputRef = useRef<HTMLInputElement>(null);
  const mountedRef = useRef(true);
  const operationIdRef = useRef(0);
  const [title, setTitle] = useState("");
  const [language, setLanguage] = useState("en");
  const [createError, setCreateError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [operationLabel, setOperationLabel] = useState<"create" | "import" | null>(null);

  useEffect(() => {
    mountedRef.current = true;

    return () => {
      mountedRef.current = false;
      operationIdRef.current += 1;
    };
  }, []);

  function isCurrentOperation(operationId: number): boolean {
    return mountedRef.current && operationIdRef.current === operationId;
  }

  function handleCreateProject(event: React.FormEvent<HTMLFormElement>): void {
    event.preventDefault();
    const trimmedTitle = title.trim();
    if (!trimmedTitle || creating) return;
    const operationId = operationIdRef.current + 1;
    operationIdRef.current = operationId;

    setCreating(true);
    setOperationLabel("create");
    setCreateError(null);
    void createProject({ title: trimmedTitle, language: language.trim() || "en" })
      .then((project) => {
        if (isCurrentOperation(operationId)) {
          navigate(`/projects/${encodeURIComponent(project.id)}`);
        }
      })
      .catch((cause: unknown) => {
        if (isCurrentOperation(operationId)) {
          setCreateError(cause instanceof Error ? cause.message : "Could not create project");
        }
      })
      .finally(() => {
        if (isCurrentOperation(operationId)) {
          setCreating(false);
          setOperationLabel(null);
        }
      });
  }

  function handleImportClick(): void {
    importInputRef.current?.click();
  }

  function handleImportProject(event: React.ChangeEvent<HTMLInputElement>): void {
    const input = event.currentTarget;
    const file = event.target.files?.[0];
    if (!file) return;
    const operationId = operationIdRef.current + 1;
    operationIdRef.current = operationId;

    setCreating(true);
    setOperationLabel("import");
    setCreateError(null);
    void importElysiumProject(file)
      .then((project) => {
        if (isCurrentOperation(operationId)) {
          navigate(`/projects/${encodeURIComponent(project.id)}`);
        }
      })
      .catch((cause: unknown) => {
        if (isCurrentOperation(operationId)) {
          setCreateError(cause instanceof Error ? cause.message : "Could not import project");
        }
      })
      .finally(() => {
        if (isCurrentOperation(operationId)) {
          setCreating(false);
          setOperationLabel(null);
          input.value = "";
        }
      });
  }

  function handleLogout(): void {
    operationIdRef.current += 1;
    clearToken();
    navigate("/login", { replace: true });
  }

  return (
    <main className="app-shell">
      <section className="mx-auto flex w-full max-w-6xl flex-col gap-8" aria-labelledby="projects-page-title">
        <header className="flex flex-col gap-4 border-b border-border pb-6 sm:flex-row sm:items-start sm:justify-between">
          <div className="min-w-0">
            <p className="mb-1 text-sm font-medium text-muted-foreground">Open Edda</p>
            <h1 id="projects-page-title" className="text-3xl font-semibold tracking-normal text-foreground">
              Projects
            </h1>
            <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
              Create a new writing workspace or open an existing story project.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button asChild variant="outline">
              <Link to="/settings">
                <Settings2 data-icon="inline-start" aria-hidden="true" />
                Settings
              </Link>
            </Button>
            <Button type="button" variant="outline" onClick={handleLogout}>
              <LogOut data-icon="inline-start" aria-hidden="true" />
              Logout
            </Button>
          </div>
        </header>

        <div className="grid gap-6 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
          <section className="rounded-lg border border-border bg-background p-5 shadow-sm" aria-labelledby="create-project-title">
            <div className="mb-5">
              <h2 id="create-project-title" className="text-lg font-semibold text-foreground">
                Start a project
              </h2>
              <p className="mt-1 text-sm text-muted-foreground">Name the workspace and set its default language.</p>
            </div>

            <form className="grid gap-4" onSubmit={handleCreateProject}>
              <label className="grid gap-1.5 text-sm font-medium text-foreground">
                Title
                <input
                  className="h-9 rounded-md border border-input bg-background px-3 text-sm text-foreground outline-none transition focus:border-ring focus:ring-3 focus:ring-ring/30"
                  value={title}
                  onChange={(event) => setTitle(event.target.value)}
                  placeholder="The Glass Archive"
                  disabled={creating}
                />
              </label>
              <label className="grid gap-1.5 text-sm font-medium text-foreground">
                Language
                <input
                  className="h-9 rounded-md border border-input bg-background px-3 text-sm text-foreground outline-none transition focus:border-ring focus:ring-3 focus:ring-ring/30"
                  value={language}
                  onChange={(event) => setLanguage(event.target.value)}
                  placeholder="en"
                  disabled={creating}
                />
              </label>
              {createError ? (
                <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">
                  {createError}
                </p>
              ) : null}
              <div className="flex flex-col gap-2 sm:flex-row">
                <Button type="submit" className="w-full sm:w-auto" disabled={!title.trim() || creating}>
                  <Plus data-icon="inline-start" aria-hidden="true" />
                  {operationLabel === "create" ? "Creating..." : "Create project"}
                </Button>
                <Button type="button" variant="outline" className="w-full sm:w-auto" onClick={handleImportClick} disabled={creating}>
                  <Upload data-icon="inline-start" aria-hidden="true" />
                  {operationLabel === "import" ? "Importing..." : "Import Elysium"}
                </Button>
                <input
                  id="project-import-input"
                  ref={importInputRef}
                  className="sr-only"
                  type="file"
                  accept=".zip,application/zip"
                  onChange={handleImportProject}
                  disabled={creating}
                />
              </div>
            </form>
          </section>

          <section className="min-w-0" aria-labelledby="story-workspaces-title">
            <div className="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <h2 id="story-workspaces-title" className="text-lg font-semibold text-foreground">
                  Story workspaces
                </h2>
                <p className="text-sm text-muted-foreground">
                  {projects.length === 1 ? "1 project available" : `${projects.length} projects available`}
                </p>
              </div>
              <Button type="button" variant="outline" onClick={reload} disabled={loading}>
                {loading ? "Loading..." : "Refresh"}
              </Button>
            </div>

            {loading ? (
              <div className="rounded-lg border border-dashed border-border bg-background p-6 text-sm text-muted-foreground" aria-live="polite">
                Loading story projects...
              </div>
            ) : error ? (
              <div className="grid gap-3 rounded-lg border border-destructive/30 bg-background p-6">
                <p className="text-sm text-destructive" role="alert">
                  Could not load story projects: {error}
                </p>
                <Button type="button" variant="outline" onClick={reload} className="w-fit">
                  Try again
                </Button>
              </div>
            ) : projects.length === 0 ? (
              <div className="rounded-lg border border-dashed border-border bg-background p-6">
                <h3 className="text-base font-semibold text-foreground">No story projects yet</h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  Create a project with the form or import an Elysium archive to begin.
                </p>
              </div>
            ) : (
              <nav className="grid gap-3 sm:grid-cols-2" aria-label="Story projects">
                {projects.map((project) => (
                  <Link
                    key={project.id}
                    to={`/projects/${encodeURIComponent(project.id)}`}
                    className="group grid min-h-36 gap-4 rounded-lg border border-border bg-background p-4 text-left shadow-sm transition hover:border-ring hover:bg-muted/40 focus:outline-none focus:ring-3 focus:ring-ring/30"
                  >
                    <span className="flex min-w-0 items-start justify-between gap-3">
                      <span className="min-w-0">
                        <span className="block truncate text-base font-semibold text-foreground">{project.title}</span>
                        <span className="mt-1 block truncate text-sm text-muted-foreground">{project.slug}</span>
                      </span>
                      <ArrowUpRight className="mt-0.5 size-4 shrink-0 text-muted-foreground transition group-hover:text-foreground" aria-hidden="true" />
                    </span>
                    <span className="flex flex-wrap items-end justify-between gap-3 text-sm">
                      <span className="rounded-md border border-border bg-muted px-2 py-1 font-medium text-foreground">
                        {project.language || "Language not set"}
                      </span>
                      <span className="text-muted-foreground">Open workspace</span>
                    </span>
                  </Link>
                ))}
              </nav>
            )}
          </section>
        </div>
      </section>
    </main>
  );
}
