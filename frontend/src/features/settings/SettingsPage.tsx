import { AlertCircle, Boxes, ChevronLeft, Loader2, RefreshCw, ServerCog } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link, useLocation } from "react-router-dom";

import { listProjects } from "../../api";
import type { AppDispatch, RootState } from "../../app/store/store";
import { Button } from "../../shared/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../shared/ui/tabs";
import type { StoryProject } from "../../types";
import { ModelSettingsPanel } from "../model-settings/ModelSettingsPanel";
import { ScriptRuntimePanel } from "../script-runtime/ScriptRuntimePanel";
import { SkillsPanel } from "../skills/SkillsPanel";
import { loadProjectSkills } from "../skills/skillsThunks";
import { settingsActions } from "./settingsSlice";

export function SettingsPage() {
  const dispatch = useDispatch<AppDispatch>();
  const location = useLocation();
  const selectedProjectId = useSelector((state: RootState) => state.settings.selectedProjectId);
  const providersStatus = useSelector((state: RootState) => state.modelSettings.providersStatus);
  const providerCount = useSelector((state: RootState) => state.modelSettings.providers.length);
  const skillsProjectId = useSelector((state: RootState) => state.skills.projectId);
  const skillsStatus = useSelector((state: RootState) => state.skills.skillsStatus);
  const [projects, setProjects] = useState<StoryProject[]>([]);
  const [projectsStatus, setProjectsStatus] = useState<"idle" | "pending" | "succeeded" | "failed">("idle");
  const [projectsError, setProjectsError] = useState<string | null>(null);
  const [projectsReloadKey, setProjectsReloadKey] = useState(0);
  const [requestedProjectId, setRequestedProjectId] = useState<string | null>(null);
  const [requestedProjectApplied, setRequestedProjectApplied] = useState(false);

  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) ?? null,
    [projects, selectedProjectId],
  );

  useEffect(() => {
    setRequestedProjectId(new URLSearchParams(location.search).get("projectId"));
    setRequestedProjectApplied(false);
  }, [location.search]);

  useEffect(() => {
    let cancelled = false;

    setProjectsStatus("pending");
    setProjectsError(null);

    void listProjects()
      .then((items) => {
        if (cancelled) return;
        setProjects(items);
        setProjectsStatus("succeeded");
      })
      .catch((cause: unknown) => {
        if (cancelled) return;
        setProjectsStatus("failed");
        setProjectsError(cause instanceof Error ? cause.message : "Could not load projects");
      });

    return () => {
      cancelled = true;
    };
  }, [projectsReloadKey]);

  useEffect(() => {
    if (projectsStatus !== "succeeded") return;

    if (projects.length === 0) {
      if (selectedProjectId) {
        dispatch(settingsActions.setSelectedProjectId(null));
      }
      return;
    }

    if (!requestedProjectApplied) {
      setRequestedProjectApplied(true);
      if (requestedProjectId && projects.some((project) => project.id === requestedProjectId)) {
        if (selectedProjectId !== requestedProjectId) {
          dispatch(settingsActions.setSelectedProjectId(requestedProjectId));
        }
        return;
      }
    }

    if (!selectedProjectId || !projects.some((project) => project.id === selectedProjectId)) {
      dispatch(settingsActions.setSelectedProjectId(projects[0].id));
    }
  }, [dispatch, projects, projectsStatus, requestedProjectApplied, requestedProjectId, selectedProjectId]);

  useEffect(() => {
    if (!selectedProject || skillsProjectId === selectedProject.id) return;
    void dispatch(loadProjectSkills({ projectId: selectedProject.id }));
  }, [dispatch, selectedProject, skillsProjectId]);

  function handleRefreshProjects(): void {
    setProjectsReloadKey((current) => current + 1);
  }

  return (
    <main className="min-h-dvh bg-background p-4 text-foreground md:p-6">
      <div className="mx-auto flex max-w-6xl flex-col gap-5">
        <header className="flex flex-col gap-3 border-b border-border pb-4 md:flex-row md:items-start md:justify-between">
          <div className="min-w-0">
            <Button asChild variant="ghost" size="sm" className="-ml-2 mb-2 w-fit">
              <Link to="/projects">
                <ChevronLeft data-icon="inline-start" aria-hidden="true" />
                Projects
              </Link>
            </Button>
            <h1 className="text-2xl font-semibold tracking-normal text-foreground">System settings</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Configure providers, model defaults, skills, and script runtime controls.
            </p>
          </div>
          <Button
            type="button"
            variant="outline"
            onClick={handleRefreshProjects}
            disabled={projectsStatus === "pending"}
          >
            {projectsStatus === "pending" ? (
              <Loader2 data-icon="inline-start" className="animate-spin" aria-hidden="true" />
            ) : (
              <RefreshCw data-icon="inline-start" aria-hidden="true" />
            )}
            Refresh
          </Button>
        </header>

        <section className="grid gap-3 md:grid-cols-3" aria-label="System status">
          <div className="rounded-md border border-border bg-muted/30 p-3">
            <p className="text-xs font-medium uppercase text-muted-foreground">Projects</p>
            <p className="mt-1 text-lg font-semibold text-foreground">
              {projectsStatus === "pending" ? "Loading" : projects.length}
            </p>
            <p className="mt-1 truncate text-sm text-muted-foreground">
              {selectedProject ? selectedProject.title : "No project selected"}
            </p>
          </div>
          <div className="rounded-md border border-border bg-muted/30 p-3">
            <p className="text-xs font-medium uppercase text-muted-foreground">Providers</p>
            <p className="mt-1 text-lg font-semibold text-foreground">
              {providersStatus === "pending" ? "Loading" : providerCount}
            </p>
            <p className="mt-1 text-sm text-muted-foreground">Provider configuration status: {providersStatus}</p>
          </div>
          <div className="rounded-md border border-border bg-muted/30 p-3">
            <p className="text-xs font-medium uppercase text-muted-foreground">Skills</p>
            <p className="mt-1 text-lg font-semibold text-foreground">
              {selectedProject ? "Project scoped" : "Unavailable"}
            </p>
            <p className="mt-1 text-sm text-muted-foreground">Skill catalog status: {skillsStatus}</p>
          </div>
        </section>

        {projectsError ? (
          <p role="alert" className="flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
            <AlertCircle className="mt-0.5 size-4" aria-hidden="true" />
            {projectsError}
          </p>
        ) : null}

        <Tabs defaultValue="providers" className="min-h-0">
          <TabsList>
            <TabsTrigger value="providers">
              <ServerCog data-icon="inline-start" aria-hidden="true" />
              Providers and models
            </TabsTrigger>
            <TabsTrigger value="skills">
              <Boxes data-icon="inline-start" aria-hidden="true" />
              Skills and scripts
            </TabsTrigger>
          </TabsList>

          <TabsContent value="providers" className="min-h-0 rounded-md border border-border bg-background p-4">
            <ModelSettingsPanel />
          </TabsContent>

          <TabsContent value="skills" className="min-h-0 rounded-md border border-border bg-background p-4">
            <section className="flex min-h-0 flex-col gap-4" aria-labelledby="project-admin-title">
              <header className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                <div className="min-w-0">
                  <h2 id="project-admin-title" className="text-base font-semibold text-foreground">
                    Project administration
                  </h2>
                  <p className="text-sm text-muted-foreground">
                    {selectedProject ? selectedProject.title : "Select a project to manage skills and scripts."}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  {projects.map((project) => (
                    <Button
                      key={project.id}
                      type="button"
                      variant={project.id === selectedProject?.id ? "secondary" : "outline"}
                      size="sm"
                      aria-pressed={project.id === selectedProject?.id}
                      onClick={() => dispatch(settingsActions.setSelectedProjectId(project.id))}
                    >
                      {project.title}
                    </Button>
                  ))}
                </div>
              </header>

              {projectsStatus === "pending" ? (
                <p className="flex items-center gap-2 rounded-md border border-border p-3 text-sm text-muted-foreground">
                  <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                  Loading projects...
                </p>
              ) : null}

              {projectsStatus !== "pending" && projects.length === 0 ? (
                <div className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
                  <p>No story projects are available for project-scoped administration.</p>
                  <Button asChild variant="outline" size="sm" className="mt-3">
                    <Link to="/projects">Open projects</Link>
                  </Button>
                </div>
              ) : null}

              {selectedProject ? (
                <div className="grid min-h-0 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)]">
                  {skillsProjectId === selectedProject.id ? (
                    <SkillsPanel projectId={selectedProject.id} sessionId={null} />
                  ) : (
                    <section className="flex flex-col gap-3" aria-labelledby="skills-panel-loading-title">
                      <div>
                        <h3 id="skills-panel-loading-title" className="text-sm font-medium text-foreground">
                          Session skills
                        </h3>
                        <p className="text-xs text-muted-foreground">Loading project skills...</p>
                      </div>
                      <p className="flex items-center gap-2 rounded-md border border-border p-3 text-sm text-muted-foreground">
                        <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                        Loading available skills...
                      </p>
                    </section>
                  )}
                  <ScriptRuntimePanel projectId={selectedProject.id} />
                </div>
              ) : null}
            </section>
          </TabsContent>
        </Tabs>
      </div>
    </main>
  );
}
