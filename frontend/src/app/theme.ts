import { alpha, createTheme, type PaletteMode } from "@mui/material/styles";

export function createAppTheme(mode: PaletteMode) {
  const isDarkMode = mode === "dark";

  return createTheme({
    palette: {
      mode,
      primary: {
        main: isDarkMode ? "#2dd4bf" : "#0f766e",
      },
      secondary: {
        main: isDarkMode ? "#60a5fa" : "#2563eb",
      },
      background: {
        default: isDarkMode ? "#0b1220" : "#f4f7fb",
        paper: isDarkMode ? "#111827" : "#ffffff",
      },
      text: {
        primary: isDarkMode ? "#e5eefc" : "#0f172a",
        secondary: isDarkMode ? "#9fb0c9" : "#475569",
      },
      divider: isDarkMode ? "rgba(148, 163, 184, 0.16)" : "rgba(15, 23, 42, 0.08)",
    },
    shape: {
      borderRadius: 18,
    },
    typography: {
      fontFamily: '"Segoe UI", "Helvetica Neue", sans-serif',
      h3: {
        fontWeight: 700,
        letterSpacing: "-0.03em",
      },
      h4: {
        fontWeight: 700,
        letterSpacing: "-0.02em",
      },
      h5: {
        fontWeight: 700,
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            background: isDarkMode
              ? [
                  "radial-gradient(circle at top left, rgba(96, 165, 250, 0.1), transparent 24%)",
                  "radial-gradient(circle at top right, rgba(45, 212, 191, 0.08), transparent 22%)",
                  "linear-gradient(180deg, #0b1220 0%, #111827 100%)",
                ].join(", ")
              : [
                  "radial-gradient(circle at top left, rgba(37, 99, 235, 0.1), transparent 28%)",
                  "radial-gradient(circle at top right, rgba(15, 118, 110, 0.12), transparent 26%)",
                  "linear-gradient(180deg, #f8fbff 0%, #eef4fb 100%)",
                ].join(", "),
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            border: `1px solid ${isDarkMode ? "rgba(148, 163, 184, 0.14)" : "rgba(15, 23, 42, 0.08)"}`,
            boxShadow: isDarkMode
              ? "0 18px 45px rgba(2, 6, 23, 0.36)"
              : "0 18px 45px rgba(15, 23, 42, 0.06)",
          },
        },
      },
      MuiDialog: {
        styleOverrides: {
          paper: {
            backgroundImage: "none",
          },
        },
      },
      MuiButton: {
        defaultProps: {
          disableElevation: true,
        },
        styleOverrides: {
          root: {
            borderRadius: 14,
            textTransform: "none",
            fontWeight: 600,
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            borderColor: alpha(isDarkMode ? "#cbd5e1" : "#0f172a", isDarkMode ? 0.16 : 0.1),
          },
        },
      },
    },
  });
}
