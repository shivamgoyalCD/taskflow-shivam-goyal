import { useState } from "react";
import {
  Alert,
  Box,
  Card,
  CardContent,
  Chip,
  Divider,
  Grid2,
  MenuItem,
  Select,
  Stack,
  type SelectChangeEvent,
  Skeleton,
  Typography,
  useTheme,
} from "@mui/material";
import RadioButtonUncheckedRoundedIcon from "@mui/icons-material/RadioButtonUncheckedRounded";
import AutorenewRoundedIcon from "@mui/icons-material/AutorenewRounded";
import CheckCircleRoundedIcon from "@mui/icons-material/CheckCircleRounded";
import { useParams } from "react-router-dom";
import type { Task } from "@/api/tasks";
import { useProjectDetailQuery, useProjectStatsQuery } from "@/features/projects/useProjectDetail";

type TaskStatus = Task["status"];

const statusSections: Array<{
  key: TaskStatus;
  label: string;
  icon: typeof RadioButtonUncheckedRoundedIcon;
}> = [
  { key: "todo", label: "Todo", icon: RadioButtonUncheckedRoundedIcon },
  { key: "in_progress", label: "In Progress", icon: AutorenewRoundedIcon },
  { key: "done", label: "Done", icon: CheckCircleRoundedIcon },
];

export function ProjectDetailsPage() {
  const { id } = useParams();
  const theme = useTheme();
  const projectId = id ?? "";

  const [statusFilter, setStatusFilter] = useState<"all" | TaskStatus>("all");
  const [assigneeFilter, setAssigneeFilter] = useState<string>("all");

  const projectQuery = useProjectDetailQuery(projectId);
  const statsQuery = useProjectStatsQuery(projectId);

  const project = projectQuery.data;
  const stats = statsQuery.data;

  const assigneeNameMap = new Map<string, string>();
  for (const assignee of stats?.assignee_counts ?? []) {
    if (assignee.assignee_id && assignee.assignee_name) {
      assigneeNameMap.set(assignee.assignee_id, assignee.assignee_name);
    }
  }

  const filteredTasks =
    project?.tasks.filter((task) => {
      const matchesStatus = statusFilter === "all" || task.status === statusFilter;
      const matchesAssignee =
        assigneeFilter === "all"
          ? true
          : assigneeFilter === "unassigned"
            ? task.assignee_id === null
            : task.assignee_id === assigneeFilter;

      return matchesStatus && matchesAssignee;
    }) ?? [];

  const tasksByStatus = {
    todo: filteredTasks.filter((task) => task.status === "todo"),
    in_progress: filteredTasks.filter((task) => task.status === "in_progress"),
    done: filteredTasks.filter((task) => task.status === "done"),
  };

  function handleStatusChange(event: SelectChangeEvent<"all" | TaskStatus>) {
    setStatusFilter(event.target.value as "all" | TaskStatus);
  }

  function handleAssigneeChange(event: SelectChangeEvent<string>) {
    setAssigneeFilter(event.target.value);
  }

  if (!projectId) {
    return (
      <Alert severity="error">
        The project ID is missing from the route. Return to the projects page and try
        again.
      </Alert>
    );
  }

  if (projectQuery.isLoading || statsQuery.isLoading) {
    return (
      <Stack spacing={4}>
        <Stack spacing={1.5}>
          <Skeleton variant="rounded" width={220} height={32} />
          <Skeleton variant="text" width="45%" height={54} />
          <Skeleton variant="text" width="72%" />
          <Skeleton variant="text" width="58%" />
        </Stack>

        <Grid2 container spacing={3}>
          {statusSections.map((section) => (
            <Grid2 key={section.key} size={{ xs: 12, sm: 4 }}>
              <Card>
                <CardContent>
                  <Stack spacing={1.5}>
                    <Skeleton variant="text" width="45%" />
                    <Skeleton variant="rounded" height={64} />
                  </Stack>
                </CardContent>
              </Card>
            </Grid2>
          ))}
        </Grid2>

        <Card>
          <CardContent sx={{ p: 3 }}>
            <Stack spacing={2}>
              <Skeleton variant="text" width="30%" />
              <Skeleton variant="rounded" height={56} />
              <Skeleton variant="rounded" height={220} />
            </Stack>
          </CardContent>
        </Card>
      </Stack>
    );
  }

  if (projectQuery.isError || statsQuery.isError) {
    return (
      <Alert severity="error">
        Unable to load this project right now. Please refresh or return to the projects
        page and try again.
      </Alert>
    );
  }

  return (
    <Stack spacing={4}>
      <Stack spacing={1.5}>
        <Chip
          label={`Project ID: ${projectId}`}
          color="secondary"
          variant="outlined"
          sx={{ width: "fit-content" }}
        />
        <Typography variant="h3">{project?.name ?? "Project"}</Typography>
        <Typography color="text.secondary" maxWidth={760}>
          {project?.description?.trim()
            ? project.description
            : "No project description has been added yet."}
        </Typography>
      </Stack>

      <Grid2 container spacing={3}>
        {statusSections.map((section) => {
          const Icon = section.icon;
          const total =
            section.key === "todo"
              ? stats?.status_counts.todo ?? 0
              : section.key === "in_progress"
                ? stats?.status_counts.in_progress ?? 0
                : stats?.status_counts.done ?? 0;

          return (
            <Grid2 key={section.key} size={{ xs: 12, sm: 4 }}>
              <Card>
                <CardContent>
                  <Stack spacing={1.5}>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Icon color="primary" fontSize="small" />
                      <Typography color="text.secondary">{section.label}</Typography>
                    </Stack>
                    <Typography variant="h4">{String(total).padStart(2, "0")}</Typography>
                    <Typography variant="body2" color="text.secondary">
                      {section.label} tasks in the project
                    </Typography>
                  </Stack>
                </CardContent>
              </Card>
            </Grid2>
          );
        })}
      </Grid2>

      <Card>
        <CardContent sx={{ p: 3 }}>
          <Stack spacing={2.5}>
            <Stack
              direction={{ xs: "column", lg: "row" }}
              spacing={2}
              justifyContent="space-between"
              alignItems={{ xs: "stretch", lg: "center" }}
            >
              <Box>
                <Typography variant="h5">Tasks</Typography>
                <Typography color="text.secondary">
                  {filteredTasks.length} visible task{filteredTasks.length === 1 ? "" : "s"} after
                  filters
                </Typography>
              </Box>

              <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
                <Select size="small" value={statusFilter} onChange={handleStatusChange} displayEmpty>
                  <MenuItem value="all">All statuses</MenuItem>
                  <MenuItem value="todo">Todo</MenuItem>
                  <MenuItem value="in_progress">In Progress</MenuItem>
                  <MenuItem value="done">Done</MenuItem>
                </Select>

                <Select
                  size="small"
                  value={assigneeFilter}
                  onChange={handleAssigneeChange}
                  displayEmpty
                >
                  <MenuItem value="all">All assignees</MenuItem>
                  <MenuItem value="unassigned">Unassigned</MenuItem>
                  {(stats?.assignee_counts ?? [])
                    .filter((assignee) => assignee.assignee_id && assignee.assignee_name)
                    .map((assignee) => (
                      <MenuItem key={assignee.assignee_id} value={assignee.assignee_id ?? ""}>
                        {assignee.assignee_name}
                      </MenuItem>
                    ))}
                </Select>
              </Stack>
            </Stack>

            <Divider />

            {project?.tasks.length === 0 ? (
              <Stack spacing={1.5} alignItems="center" sx={{ py: 6, textAlign: "center" }}>
                <Typography variant="h6">No tasks yet</Typography>
                <Typography color="text.secondary" maxWidth={560}>
                  This project does not have any tasks yet. When tasks are created, they will
                  appear here grouped by status.
                </Typography>
              </Stack>
            ) : filteredTasks.length === 0 ? (
              <Stack spacing={1.5} alignItems="center" sx={{ py: 6, textAlign: "center" }}>
                <Typography variant="h6">No tasks match these filters</Typography>
                <Typography color="text.secondary" maxWidth={560}>
                  Change the selected status or assignee filter to see more tasks.
                </Typography>
              </Stack>
            ) : (
              <Grid2 container spacing={3}>
                {statusSections
                  .filter((section) => statusFilter === "all" || section.key === statusFilter)
                  .map((section) => {
                    const Icon = section.icon;
                    const sectionTasks = tasksByStatus[section.key];

                    return (
                      <Grid2 key={section.key} size={{ xs: 12, xl: 4 }}>
                        <Card
                          sx={{
                            height: "100%",
                            backgroundColor:
                              theme.palette.mode === "dark"
                                ? "rgba(15, 23, 42, 0.36)"
                                : "rgba(255, 255, 255, 0.72)",
                          }}
                        >
                          <CardContent sx={{ p: 2.5 }}>
                            <Stack spacing={2}>
                              <Stack
                                direction="row"
                                spacing={1}
                                justifyContent="space-between"
                                alignItems="center"
                              >
                                <Stack direction="row" spacing={1} alignItems="center">
                                  <Icon color="primary" fontSize="small" />
                                  <Typography variant="h6">{section.label}</Typography>
                                </Stack>
                                <Chip
                                  size="small"
                                  color="primary"
                                  label={`${sectionTasks.length} task${sectionTasks.length === 1 ? "" : "s"}`}
                                />
                              </Stack>

                              {sectionTasks.length === 0 ? (
                                <Box
                                  sx={{
                                    minHeight: 120,
                                    display: "grid",
                                    placeItems: "center",
                                    borderRadius: 3,
                                    border: `1px dashed ${theme.palette.divider}`,
                                  }}
                                >
                                  <Typography variant="body2" color="text.secondary">
                                    No tasks in this group
                                  </Typography>
                                </Box>
                              ) : (
                                <Stack spacing={1.5}>
                                  {sectionTasks.map((task) => (
                                    <Card
                                      key={task.id}
                                      variant="outlined"
                                      sx={{
                                        borderRadius: 3,
                                        backgroundColor:
                                          theme.palette.mode === "dark"
                                            ? "rgba(17, 24, 39, 0.88)"
                                            : "rgba(255, 255, 255, 0.96)",
                                      }}
                                    >
                                      <CardContent sx={{ p: 2 }}>
                                        <Stack spacing={1.5}>
                                          <Stack
                                            direction="row"
                                            spacing={1}
                                            justifyContent="space-between"
                                            alignItems="flex-start"
                                          >
                                            <Box>
                                              <Typography fontWeight={700}>{task.title}</Typography>
                                              <Typography variant="caption" color="text.secondary">
                                                Updated {new Date(task.updated_at).toLocaleDateString()}
                                              </Typography>
                                            </Box>
                                            <Chip
                                              size="small"
                                              variant="outlined"
                                              color={priorityColor(task.priority)}
                                              label={priorityLabel(task.priority)}
                                            />
                                          </Stack>

                                          <Typography variant="body2" color="text.secondary">
                                            {task.description?.trim()
                                              ? task.description
                                              : "No task description provided."}
                                          </Typography>

                                          <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                                            <Chip
                                              size="small"
                                              variant="outlined"
                                              label={
                                                task.assignee_id
                                                  ? assigneeNameMap.get(task.assignee_id) ??
                                                    `User ${task.assignee_id.slice(0, 8)}...`
                                                  : "Unassigned"
                                              }
                                            />
                                            <Chip
                                              size="small"
                                              variant="outlined"
                                              label={
                                                task.due_date
                                                  ? `Due ${new Date(task.due_date).toLocaleDateString()}`
                                                  : "No due date"
                                              }
                                            />
                                          </Stack>
                                        </Stack>
                                      </CardContent>
                                    </Card>
                                  ))}
                                </Stack>
                              )}
                            </Stack>
                          </CardContent>
                        </Card>
                      </Grid2>
                    );
                  })}
              </Grid2>
            )}
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ p: 3 }}>
          <Stack spacing={2}>
            <Typography variant="h5">Assignee distribution</Typography>
            <Divider />
            {stats?.assignee_counts.length ? (
              <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                {stats.assignee_counts.map((assignee) => (
                  <Chip
                    key={assignee.assignee_id ?? "unassigned"}
                    label={`${assignee.assignee_name ?? "Unassigned"}: ${assignee.count}`}
                    variant="outlined"
                  />
                ))}
              </Stack>
            ) : (
              <Typography color="text.secondary">
                No assignee summary is available yet.
              </Typography>
            )}
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}

function priorityLabel(priority: Task["priority"]) {
  if (priority === "high") {
    return "High priority";
  }
  if (priority === "medium") {
    return "Medium priority";
  }
  return "Low priority";
}

function priorityColor(
  priority: Task["priority"],
): "default" | "success" | "warning" | "error" {
  if (priority === "high") {
    return "error";
  }
  if (priority === "medium") {
    return "warning";
  }
  return "success";
}
