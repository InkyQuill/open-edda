import { useCallback, useEffect, useState } from "react";

import { listProjects } from "../../api";
import type { StoryProject } from "../../types";

export type UseProjectsResult = {
  projects: StoryProject[];
  loading: boolean;
  error: string | null;
  reload: () => void;
};

export function useProjects(): UseProjectsResult {
  const [projects, setProjects] = useState<StoryProject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadKey, setReloadKey] = useState(0);

  const reload = useCallback(() => {
    setReloadKey((current) => current + 1);
  }, []);

  useEffect(() => {
    let cancelled = false;

    setLoading(true);
    setError(null);

    void listProjects()
      .then((items) => {
        if (cancelled) {
          return;
        }
        setProjects(items);
        setError(null);
      })
      .catch((cause: unknown) => {
        if (cancelled) {
          return;
        }
        setProjects([]);
        setError(cause instanceof Error ? cause.message : "Project list failed");
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [reloadKey]);

  return { projects, loading, error, reload };
}
