import {
  Card,
  CardContent,
  Chip,
  Grid2,
  Stack,
  Typography,
} from "@mui/material";
import { Link as RouterLink } from "react-router-dom";

const placeholderProjects = [
  {
    id: "alpha",
    name: "Client Launch",
    description: "Placeholder project card until API integration is added.",
    tasks: 12,
    completed: 5,
  },
  {
    id: "beta",
    name: "Internal Planning",
    description: "Route shell for project dashboards and stats.",
    tasks: 8,
    completed: 3,
  },
];

export function ProjectsPage() {
  return (
    <Stack spacing={4}>
      <Stack spacing={1}>
        <Typography variant="h3">Projects</Typography>
        <Typography color="text.secondary" maxWidth={720}>
          This page is wired into the app router and layout, with placeholder cards ready
          to swap over to TanStack Query once API hooks are added.
        </Typography>
      </Stack>

      <Grid2 container spacing={3}>
        {placeholderProjects.map((project) => (
          <Grid2 key={project.id} size={{ xs: 12, md: 6 }}>
            <Card
              component={RouterLink}
              to={`/projects/${project.id}`}
              sx={{
                display: "block",
                transition: "transform 160ms ease, box-shadow 160ms ease",
                "&:hover": {
                  transform: "translateY(-3px)",
                  boxShadow: "0 24px 44px rgba(15, 23, 42, 0.1)",
                },
              }}
            >
              <CardContent sx={{ p: 3 }}>
                <Stack spacing={2}>
                  <Stack
                    direction="row"
                    spacing={1}
                    justifyContent="space-between"
                    alignItems="center"
                  >
                    <Typography variant="h5">{project.name}</Typography>
                    <Chip color="primary" label={`${project.tasks} tasks`} />
                  </Stack>

                  <Typography color="text.secondary">{project.description}</Typography>

                  <Stack direction="row" spacing={1}>
                    <Chip variant="outlined" label={`${project.completed} completed`} />
                    <Chip variant="outlined" label="Frontend placeholder" />
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
          </Grid2>
        ))}
      </Grid2>
    </Stack>
  );
}
