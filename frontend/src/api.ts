import type { StoryProject } from "./types";

export async function listProjects(): Promise<StoryProject[]> {
  const response = await fetch("/api/projects");
  if (!response.ok) {
    throw new Error(`list projects failed: ${response.status}`);
  }
  return response.json() as Promise<StoryProject[]>;
}
