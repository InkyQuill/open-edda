import { Navigate, Route, Routes } from "react-router-dom";
import { AuthPage } from "../../features/auth/AuthPage";
import { ProjectsPage } from "../../features/projects/ProjectsPage";
import { WorkspacePage } from "../../features/workspace/WorkspacePage";
import { RequireAuth } from "./RequireAuth";

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<AuthPage />} />
      <Route element={<RequireAuth />}>
        <Route path="/projects" element={<ProjectsPage />} />
        <Route path="/projects/:projectId" element={<WorkspacePage />} />
        <Route
          path="/projects/:projectId/content/:contentKind/:contentId"
          element={<WorkspacePage />}
        />
        <Route
          path="/projects/:projectId/content/:contentId"
          element={<WorkspacePage />}
        />
      </Route>
      <Route path="/" element={<Navigate to="/projects" replace />} />
      <Route path="*" element={<Navigate to="/projects" replace />} />
    </Routes>
  );
}
