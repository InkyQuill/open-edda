import { useEffect, useState } from "react";
import type { StoryProject } from "./types";
import { listProjects } from "./api";
import "./styles.css";

export function App() {
  const [projects, setProjects] = useState<StoryProject[]>([]);

  useEffect(() => {
    void listProjects().then(setProjects);
  }, []);

  return (
    <main className="app-shell">
      <section className="project-dashboard">
        <header>
          <h1>Writer</h1>
          <p>Private AI writing studio</p>
        </header>

        <div className="project-list">
          {projects.length === 0 ? (
            <p>No story projects yet.</p>
          ) : (
            projects.map((project) => (
              <article key={project.id} className="project-row">
                <h2>{project.title}</h2>
                <p>{project.language || "Language not set"}</p>
              </article>
            ))
          )}
        </div>
      </section>
    </main>
  );
}
