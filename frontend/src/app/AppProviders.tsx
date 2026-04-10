import type { PropsWithChildren } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { queryClient } from "@/app/queryClient";
import { theme } from "@/app/theme";
import { AuthProvider } from "@/features/auth/AuthContext";

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <AuthProvider>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider theme={theme}>
          <CssBaseline />
          {children}
        </ThemeProvider>
      </QueryClientProvider>
    </AuthProvider>
  );
}
