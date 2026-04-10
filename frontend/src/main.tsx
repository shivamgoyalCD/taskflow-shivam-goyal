import React from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider } from "react-router-dom";
import { AppProviders } from "@/app/AppProviders";
import { RouteLoadingState } from "@/components/RouteLoadingState";
import { router } from "@/routes/router";
import "@/app/styles.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <AppProviders>
      <RouterProvider router={router} fallbackElement={<RouteLoadingState />} />
    </AppProviders>
  </React.StrictMode>,
);
