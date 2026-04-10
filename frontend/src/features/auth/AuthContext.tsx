import {
  createContext,
  useContext,
  useEffect,
  useState,
  type PropsWithChildren,
} from "react";
import { clearStoredSession, loadStoredSession, persistSession } from "@/features/auth/authStorage";
import type { AuthSession, AuthUser } from "@/types/auth";

type AuthContextValue = {
  isHydrated: boolean;
  isAuthenticated: boolean;
  token: string | null;
  user: AuthUser | null;
  setSession: (session: AuthSession) => void;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSessionState] = useState<AuthSession | null>(() => loadStoredSession());
  const [isHydrated, setIsHydrated] = useState(false);

  useEffect(() => {
    setSessionState(loadStoredSession());
    setIsHydrated(true);
  }, []);

  function setSession(nextSession: AuthSession) {
    persistSession(nextSession);
    setSessionState(nextSession);
  }

  function logout() {
    clearStoredSession();
    setSessionState(null);
  }

  return (
    <AuthContext.Provider
      value={{
        isHydrated,
        isAuthenticated: Boolean(session),
        token: session?.token ?? null,
        user: session?.user ?? null,
        setSession,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }

  return context;
}
