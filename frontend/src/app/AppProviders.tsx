import type { PropsWithChildren } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/app/queryClient";
import { ThemeModeProvider } from "@/app/ThemeModeProvider";
import { AuthProvider } from "@/features/auth/AuthContext";

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <AuthProvider>
      <QueryClientProvider client={queryClient}>
        <ThemeModeProvider>
          {children}
        </ThemeModeProvider>
      </QueryClientProvider>
    </AuthProvider>
  );
}
