import type { AuthSession } from "@/types/auth";

const authStorageKey = "taskflow.auth";

export function loadStoredSession(): AuthSession | null {
  if (typeof window === "undefined") {
    return null;
  }

  const rawValue = window.localStorage.getItem(authStorageKey);
  if (!rawValue) {
    return null;
  }

  try {
    const parsed = JSON.parse(rawValue) as Partial<AuthSession>;
    if (
      typeof parsed.token !== "string" ||
      !parsed.user ||
      typeof parsed.user.id !== "string" ||
      typeof parsed.user.name !== "string" ||
      typeof parsed.user.email !== "string"
    ) {
      window.localStorage.removeItem(authStorageKey);
      return null;
    }

    return {
      token: parsed.token,
      user: parsed.user,
    };
  } catch {
    window.localStorage.removeItem(authStorageKey);
    return null;
  }
}

export function persistSession(session: AuthSession) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(authStorageKey, JSON.stringify(session));
}

export function clearStoredSession() {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(authStorageKey);
}
