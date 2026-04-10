import { Outlet, useNavigate } from "react-router-dom";
import {
  AppBar,
  Avatar,
  Box,
  Button,
  Container,
  Divider,
  IconButton,
  Stack,
  Tooltip,
  Toolbar,
  Typography,
  useTheme,
} from "@mui/material";
import DashboardRoundedIcon from "@mui/icons-material/DashboardRounded";
import LoginRoundedIcon from "@mui/icons-material/LoginRounded";
import LogoutRoundedIcon from "@mui/icons-material/LogoutRounded";
import DarkModeRoundedIcon from "@mui/icons-material/DarkModeRounded";
import LightModeRoundedIcon from "@mui/icons-material/LightModeRounded";
import { AppNavLink } from "@/components/AppNavLink";
import { useAuth } from "@/features/auth/AuthContext";
import { useThemeMode } from "@/app/ThemeModeProvider";

export function AppLayout() {
  const navigate = useNavigate();
  const theme = useTheme();
  const { mode, toggleMode } = useThemeMode();
  const { isAuthenticated, user, logout } = useAuth();

  function handleLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  return (
    <Box sx={{ minHeight: "100vh" }}>
      <AppBar
        position="sticky"
        color="transparent"
        elevation={0}
        sx={{
          backdropFilter: "blur(18px)",
          borderBottom: `1px solid ${theme.palette.divider}`,
          backgroundColor:
            theme.palette.mode === "dark"
              ? "rgba(11, 18, 32, 0.82)"
              : "rgba(248, 251, 255, 0.8)",
        }}
      >
        <Container maxWidth="lg">
          <Toolbar disableGutters sx={{ minHeight: 76 }}>
            <Stack direction="row" spacing={1.5} alignItems="center" flexGrow={1}>
              <Box
                sx={{
                  display: "grid",
                  placeItems: "center",
                  width: 40,
                  height: 40,
                  borderRadius: 3,
                  background: "linear-gradient(135deg, #0f766e 0%, #2563eb 100%)",
                  color: "common.white",
                }}
              >
                <DashboardRoundedIcon fontSize="small" />
              </Box>
              <Box>
                <Typography variant="h6" fontWeight={700}>
                  Taskflow
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Assignment frontend scaffold
                </Typography>
              </Box>
            </Stack>

            <Stack direction="row" spacing={1} alignItems="center">
              <Tooltip title={mode === "dark" ? "Switch to light mode" : "Switch to dark mode"}>
                <IconButton
                  color="inherit"
                  onClick={toggleMode}
                  sx={{
                    border: `1px solid ${theme.palette.divider}`,
                    backgroundColor:
                      theme.palette.mode === "dark"
                        ? "rgba(148, 163, 184, 0.08)"
                        : "rgba(255, 255, 255, 0.72)",
                  }}
                >
                  {mode === "dark" ? <LightModeRoundedIcon /> : <DarkModeRoundedIcon />}
                </IconButton>
              </Tooltip>

              {isAuthenticated ? (
                <>
                  <AppNavLink to="/projects" label="Projects" />
                  <Divider orientation="vertical" flexItem sx={{ mx: 0.5 }} />
                  <Stack direction="row" spacing={1.25} alignItems="center" sx={{ pl: 0.5 }}>
                    <Avatar
                      sx={{
                        width: 34,
                        height: 34,
                        bgcolor: "primary.main",
                        fontSize: 14,
                        fontWeight: 700,
                      }}
                    >
                      {user?.name.slice(0, 1).toUpperCase() ?? "U"}
                    </Avatar>
                    <Box sx={{ display: { xs: "none", sm: "block" } }}>
                      <Typography variant="body2" fontWeight={600}>
                        {user?.name ?? "User"}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {user?.email ?? "Signed in"}
                      </Typography>
                    </Box>
                    <Button
                      variant="outlined"
                      color="inherit"
                      startIcon={<LogoutRoundedIcon />}
                      onClick={handleLogout}
                    >
                      Logout
                    </Button>
                  </Stack>
                </>
              ) : (
                <>
                  <AppNavLink to="/register" label="Register" />
                  <Button
                    component={AppNavLink}
                    to="/login"
                    variant="contained"
                    startIcon={<LoginRoundedIcon />}
                  >
                    Login
                  </Button>
                </>
              )}
            </Stack>
          </Toolbar>
        </Container>
      </AppBar>

      <Container maxWidth="lg" sx={{ py: { xs: 4, md: 6 } }}>
        <Outlet />
      </Container>
    </Box>
  );
}
