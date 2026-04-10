import { Outlet } from "react-router-dom";
import {
  AppBar,
  Box,
  Button,
  Container,
  Stack,
  Toolbar,
  Typography,
} from "@mui/material";
import DashboardRoundedIcon from "@mui/icons-material/DashboardRounded";
import LoginRoundedIcon from "@mui/icons-material/LoginRounded";
import { AppNavLink } from "@/components/AppNavLink";

export function AppLayout() {
  return (
    <Box sx={{ minHeight: "100vh" }}>
      <AppBar
        position="sticky"
        color="transparent"
        elevation={0}
        sx={{
          backdropFilter: "blur(18px)",
          borderBottom: "1px solid rgba(15, 23, 42, 0.08)",
          backgroundColor: "rgba(248, 251, 255, 0.8)",
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
              <AppNavLink to="/projects" label="Projects" />
              <AppNavLink to="/register" label="Register" />
              <Button
                component={AppNavLink}
                to="/login"
                variant="contained"
                startIcon={<LoginRoundedIcon />}
              >
                Login
              </Button>
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
