import { createBrowserRouter, Navigate } from "react-router-dom";
import { ProtectedRoute } from "@/features/auth/ProtectedRoute";
import { AppLayout } from "@/layouts/AppLayout";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <AppLayout />,
    children: [
      {
        index: true,
        element: <Navigate to="/projects" replace />,
      },
      {
        path: "login",
        lazy: async () => {
          const module = await import("@/pages/LoginPage");
          return { Component: module.LoginPage };
        },
      },
      {
        path: "register",
        lazy: async () => {
          const module = await import("@/pages/RegisterPage");
          return { Component: module.RegisterPage };
        },
      },
      {
        path: "projects",
        element: <ProtectedRoute />,
        children: [
          {
            index: true,
            lazy: async () => {
              const module = await import("@/pages/ProjectsPage");
              return { Component: module.ProjectsPage };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const module = await import("@/pages/ProjectDetailsPage");
              return { Component: module.ProjectDetailsPage };
            },
          },
        ],
      },
      {
        path: "*",
        lazy: async () => {
          const module = await import("@/pages/NotFoundPage");
          return { Component: module.NotFoundPage };
        },
      },
    ],
  },
]);
