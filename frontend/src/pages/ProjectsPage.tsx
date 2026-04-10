import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Grid2,
  MenuItem,
  Select,
  Stack,
  type SelectChangeEvent,
  Typography,
  useTheme,
} from "@mui/material";
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import RefreshRoundedIcon from "@mui/icons-material/RefreshRounded";
import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import ArrowForwardRoundedIcon from "@mui/icons-material/ArrowForwardRounded";
import { useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import { CreateProjectDialog } from "@/features/projects/CreateProjectDialog";
import { useProjectsQuery } from "@/features/projects/useProjects";

const pageSizeOptions = [6, 12, 18] as const;

export function ProjectsPage() {
  const theme = useTheme();
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState<number>(pageSizeOptions[0]);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);

  const projectsQuery = useProjectsQuery(page, limit);
  const projects = projectsQuery.data?.projects ?? [];
  const hasProjects = projects.length > 0;
  const canGoToPreviousPage = page > 1;
  const canGoToNextPage = hasProjects && projects.length >= limit;

  const pageSummary = useMemo(
    () =>
      projectsQuery.isLoading
        ? "Loading projects..."
        : `${projects.length} project${projects.length === 1 ? "" : "s"} on this page`,
    [projects.length, projectsQuery.isLoading],
  );

  function handlePageSizeChange(event: SelectChangeEvent<number>) {
    setLimit(Number(event.target.value));
    setPage(1);
  }

  function handleCreateSuccess() {
    setPage(1);
  }

  return (
    <Stack spacing={4}>
      <Stack
        spacing={{ xs: 2, md: 1.5 }}
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        alignItems={{ xs: "stretch", md: "flex-end" }}
      >
        <Box>
          <Typography variant="h3">Projects</Typography>
          <Typography color="text.secondary" maxWidth={720}>
            Browse your accessible projects, paginate through results, and create a new
            project without leaving the page.
          </Typography>
        </Box>

        <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
          <Button
            variant="outlined"
            color="inherit"
            startIcon={<RefreshRoundedIcon />}
            onClick={() => {
              void projectsQuery.refetch();
            }}
            disabled={projectsQuery.isFetching}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AddRoundedIcon />}
            onClick={() => setIsCreateDialogOpen(true)}
          >
            Create project
          </Button>
        </Stack>
      </Stack>

      <Stack
        direction={{ xs: "column", md: "row" }}
        spacing={2}
        justifyContent="space-between"
        alignItems={{ xs: "stretch", md: "center" }}
      >
        <Typography color="text.secondary">{pageSummary}</Typography>

        <Stack direction="row" spacing={1.5} alignItems="center">
          <Typography variant="body2" color="text.secondary">
            Per page
          </Typography>
          <Select<number> size="small" value={limit} onChange={handlePageSizeChange}>
            {pageSizeOptions.map((option) => (
              <MenuItem key={option} value={option}>
                {option}
              </MenuItem>
            ))}
          </Select>
        </Stack>
      </Stack>

      {projectsQuery.isLoading ? (
        <Card>
          <CardContent sx={{ py: 8 }}>
            <Stack spacing={2} alignItems="center">
              <CircularProgress />
              <Typography color="text.secondary">Loading projects...</Typography>
            </Stack>
          </CardContent>
        </Card>
      ) : null}

      {projectsQuery.isError ? (
        <Alert
          severity="error"
          action={
            <Button
              color="inherit"
              size="small"
              onClick={() => {
                void projectsQuery.refetch();
              }}
            >
              Retry
            </Button>
          }
        >
          Unable to load projects right now. Please try again.
        </Alert>
      ) : null}

      {!projectsQuery.isLoading && !projectsQuery.isError && !hasProjects ? (
        <Card>
          <CardContent sx={{ py: 8 }}>
            <Stack spacing={2} alignItems="center" textAlign="center">
              <Typography variant="h5">No projects yet</Typography>
              <Typography color="text.secondary" maxWidth={560}>
                Create your first project to start organizing tasks, members, and status
                updates.
              </Typography>
              <Button
                variant="contained"
                startIcon={<AddRoundedIcon />}
                onClick={() => setIsCreateDialogOpen(true)}
              >
                Create project
              </Button>
            </Stack>
          </CardContent>
        </Card>
      ) : null}

      {!projectsQuery.isLoading && !projectsQuery.isError && hasProjects ? (
        <Grid2 container spacing={3}>
          {projects.map((project) => (
            <Grid2 key={project.id} size={{ xs: 12, md: 6 }}>
              <Card
                component={RouterLink}
                to={`/projects/${project.id}`}
                sx={{
                  display: "block",
                  height: "100%",
                  transition: "transform 160ms ease, box-shadow 160ms ease",
                  "&:hover": {
                    transform: "translateY(-3px)",
                    boxShadow:
                      theme.palette.mode === "dark"
                        ? "0 24px 44px rgba(2, 6, 23, 0.42)"
                        : "0 24px 44px rgba(15, 23, 42, 0.1)",
                  },
                }}
              >
                <CardContent sx={{ p: 3 }}>
                  <Stack spacing={2}>
                    <Stack
                      direction="row"
                      spacing={1}
                      justifyContent="space-between"
                      alignItems="flex-start"
                    >
                      <Stack spacing={0.5}>
                        <Typography variant="h5">{project.name}</Typography>
                        <Typography variant="body2" color="text.secondary">
                          Created {new Date(project.created_at).toLocaleDateString()}
                        </Typography>
                      </Stack>
                      <Chip color="primary" label="Project" />
                    </Stack>

                    <Typography color="text.secondary">
                      {project.description?.trim()
                        ? project.description
                        : "No description added yet."}
                    </Typography>

                    <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                      <Chip variant="outlined" label={`Owner ${project.owner_id.slice(0, 8)}...`} />
                      <Chip variant="outlined" label={`ID ${project.id.slice(0, 8)}...`} />
                    </Stack>
                  </Stack>
                </CardContent>
              </Card>
            </Grid2>
          ))}
        </Grid2>
      ) : null}

      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1.5}
        justifyContent="space-between"
        alignItems={{ xs: "stretch", sm: "center" }}
      >
        <Typography color="text.secondary">Page {page}</Typography>

        <Stack direction="row" spacing={1.5}>
          <Button
            variant="outlined"
            color="inherit"
            startIcon={<ArrowBackRoundedIcon />}
            disabled={!canGoToPreviousPage || projectsQuery.isFetching}
            onClick={() => setPage((currentPage) => Math.max(1, currentPage - 1))}
          >
            Previous
          </Button>
          <Button
            variant="outlined"
            endIcon={<ArrowForwardRoundedIcon />}
            disabled={!canGoToNextPage || projectsQuery.isFetching}
            onClick={() => setPage((currentPage) => currentPage + 1)}
          >
            Next
          </Button>
        </Stack>
      </Stack>

      <CreateProjectDialog
        open={isCreateDialogOpen}
        onClose={() => setIsCreateDialogOpen(false)}
        onCreated={handleCreateSuccess}
      />
    </Stack>
  );
}
