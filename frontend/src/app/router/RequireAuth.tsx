import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useSyncExternalStore } from "react";
import { getToken, subscribeSession, wasSessionExpired } from "../../authApi";

export function RequireAuth() {
  const location = useLocation();

  const token = useSyncExternalStore(subscribeSession, getToken, () => null);

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location.pathname + location.search + location.hash, sessionExpired: wasSessionExpired() }} />;
  }

  return <Outlet />;
}
