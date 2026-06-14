import { Link, useNavigate } from "react-router-dom";

import { clearToken } from "../../authApi";
import { Button } from "../../shared/ui/button";
import { useProjects } from "./projectHooks";

export function ProjectsPage() {
  const navigate = useNavigate();
  const { projects, loading, error, reload } = useProjects();

  function handleLogout(): void {
    clearToken();
    navigate("/login", { replace: true });
  }

  return (
    <main className="app-shell">
      <section className="project-dashboard" aria-labelledby="projects-page-title">
        <header>
          <div>
            <h1 id="projects-page-title">Projects</h1>
            <p>Choose a writing workspace.</p>
          </div>
          <Button type="button" variant="outline" onClick={handleLogout}>
            Logout
          </Button>
        </header>

        {loading ? (
          <div className="project-list" aria-live="polite">
            <p>Loading story projects...</p>
          </div>
        ) : error ? (
          <div className="project-list">
            <p role="alert">Could not load story projects: {error}</p>
            <Button type="button" variant="outline" onClick={reload}>
              Try again
            </Button>
          </div>
        ) : projects.length === 0 ? (
          <div className="project-list">
            <p>No story projects yet.</p>
            <Button type="button" variant="outline" onClick={reload}>
              Refresh
            </Button>
          </div>
        ) : (
          <nav className="project-list" aria-label="Story projects">
            {projects.map((project) => (
              <Button key={project.id} asChild variant="outline" className="project-row">
                <Link to={`/projects/${encodeURIComponent(project.id)}`}>
                  <span className="project-row-title">{project.title}</span>
                  <span className="project-row-meta">{project.language || "Language not set"}</span>
                </Link>
              </Button>
            ))}
          </nav>
        )}
      </section>
    </main>
  );
}
