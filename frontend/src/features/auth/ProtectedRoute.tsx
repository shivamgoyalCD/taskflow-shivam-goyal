import { Navigate, Outlet, useLocation } from "react-router-dom";
import { RouteLoadingState } from "@/components/RouteLoadingState";
import { useAuth } from "@/features/auth/AuthContext";

export function ProtectedRoute() {
  const location = useLocation();
  const { isAuthenticated, isHydrated } = useAuth();

  if (!isHydrated) {
    return <RouteLoadingState />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return <Outlet />;
}
