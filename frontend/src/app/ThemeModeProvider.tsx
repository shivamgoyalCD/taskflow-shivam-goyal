import {
  createContext,
  useContext,
  useEffect,
  useState,
  type PropsWithChildren,
} from "react";
import { CssBaseline, ThemeProvider, type PaletteMode } from "@mui/material";
import { createAppTheme } from "@/app/theme";

const themeStorageKey = "taskflow.theme-mode";

type ThemeModeContextValue = {
  mode: PaletteMode;
  toggleMode: () => void;
};

const ThemeModeContext = createContext<ThemeModeContextValue | undefined>(undefined);

export function ThemeModeProvider({ children }: PropsWithChildren) {
  const [mode, setMode] = useState<PaletteMode>(() => loadStoredMode());

  useEffect(() => {
    window.localStorage.setItem(themeStorageKey, mode);
    document.documentElement.dataset.theme = mode;
    document.documentElement.style.colorScheme = mode;
  }, [mode]);

  const theme = createAppTheme(mode);

  function toggleMode() {
    setMode((currentMode) => (currentMode === "light" ? "dark" : "light"));
  }

  return (
    <ThemeModeContext.Provider value={{ mode, toggleMode }}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </ThemeModeContext.Provider>
  );
}

export function useThemeMode() {
  const context = useContext(ThemeModeContext);
  if (!context) {
    throw new Error("useThemeMode must be used within ThemeModeProvider");
  }

  return context;
}

function loadStoredMode(): PaletteMode {
  if (typeof window === "undefined") {
    return "light";
  }

  const storedMode = window.localStorage.getItem(themeStorageKey);
  if (storedMode === "light" || storedMode === "dark") {
    return storedMode;
  }

  return "light";
}
