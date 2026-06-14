import { Navigate, Outlet, useLocation } from "react-router-dom";
import { getToken } from "../../authApi";

export function RequireAuth() {
  const location = useLocation();

  if (!getToken()) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return <Outlet />;
}
