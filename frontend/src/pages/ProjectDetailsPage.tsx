import { useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  Grid2,
  IconButton,
  MenuItem,
  Select,
  Snackbar,
  Stack,
  type SelectChangeEvent,
  Skeleton,
  Typography,
  useTheme,
} from "@mui/material";
import MuiAlert from "@mui/material/Alert";
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import AutorenewRoundedIcon from "@mui/icons-material/AutorenewRounded";
import CheckCircleRoundedIcon from "@mui/icons-material/CheckCircleRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import RadioButtonUncheckedRoundedIcon from "@mui/icons-material/RadioButtonUncheckedRounded";
import { useParams } from "react-router-dom";
import { ApiError } from "@/api/client";
import type { Task } from "@/api/tasks";
import { useAuth } from "@/features/auth/AuthContext";
import { useProjectDetailQuery, useProjectStatsQuery } from "@/features/projects/useProjectDetail";
import { DeleteTaskDialog } from "@/features/tasks/DeleteTaskDialog";
import {
  defaultTaskFormValues,
  type TaskFormValues,
} from "@/features/tasks/taskSchemas";
import { TaskDialog, type TaskAssigneeOption } from "@/features/tasks/TaskDialog";
import {
  useCreateTaskMutation,
  useDeleteTaskMutation,
  useUpdateTaskMutation,
} from "@/features/tasks/useTaskMutations";

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
  const { user } = useAuth();
  const projectId = id ?? "";

  const [statusFilter, setStatusFilter] = useState<"all" | TaskStatus>("all");
  const [assigneeFilter, setAssigneeFilter] = useState("all");
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [deletingTask, setDeletingTask] = useState<Task | null>(null);
  const [createApiError, setCreateApiError] = useState<string | null>(null);
  const [editApiError, setEditApiError] = useState<string | null>(null);
  const [deleteApiError, setDeleteApiError] = useState<string | null>(null);
  const [createFieldErrors, setCreateFieldErrors] = useState<
    Partial<Record<keyof TaskFormValues, string>>
  >({});
  const [editFieldErrors, setEditFieldErrors] = useState<
    Partial<Record<keyof TaskFormValues, string>>
  >({});
  const [snackbarState, setSnackbarState] = useState<{
    open: boolean;
    severity: "success" | "error";
    message: string;
  }>({ open: false, severity: "success", message: "" });

  const projectQuery = useProjectDetailQuery(projectId);
  const statsQuery = useProjectStatsQuery(projectId);
  const createTaskMutation = useCreateTaskMutation(projectId);
  const updateTaskMutation = useUpdateTaskMutation(projectId);
  const deleteTaskMutation = useDeleteTaskMutation(projectId);

  const project = projectQuery.data;
  const stats = statsQuery.data;

  const assigneeNameMap = useMemo(() => {
    const map = new Map<string, string>();
    if (user) {
      map.set(user.id, user.name);
    }
    for (const assignee of stats?.assignee_counts ?? []) {
      if (assignee.assignee_id) {
        map.set(
          assignee.assignee_id,
          assignee.assignee_name ?? `User ${assignee.assignee_id.slice(0, 8)}...`,
        );
      }
    }
    for (const task of project?.tasks ?? []) {
      if (task.assignee_id && !map.has(task.assignee_id)) {
        map.set(task.assignee_id, `User ${task.assignee_id.slice(0, 8)}...`);
      }
    }
    return map;
  }, [project?.tasks, stats?.assignee_counts, user]);

  const assigneeOptions = useMemo<TaskAssigneeOption[]>(() => {
    const namedOptions = Array.from(assigneeNameMap.entries())
      .map(([value, label]) => ({ value, label }))
      .sort((left, right) => left.label.localeCompare(right.label));

    return [{ value: "unassigned", label: "Unassigned" }, ...namedOptions];
  }, [assigneeNameMap]);

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

  function openSnackbar(severity: "success" | "error", message: string) {
    setSnackbarState({ open: true, severity, message });
  }

  function closeCreateDialog() {
    if (createTaskMutation.isPending) return;
    setIsCreateDialogOpen(false);
    setCreateApiError(null);
    setCreateFieldErrors({});
  }

  function closeEditDialog() {
    if (updateTaskMutation.isPending) return;
    setEditingTask(null);
    setEditApiError(null);
    setEditFieldErrors({});
  }

  function closeDeleteDialog() {
    if (deleteTaskMutation.isPending) return;
    setDeletingTask(null);
    setDeleteApiError(null);
  }

  async function handleCreateTask(values: TaskFormValues) {
    setCreateApiError(null);
    setCreateFieldErrors({});
    try {
      await createTaskMutation.mutateAsync(normalizeTaskFormValues(values));
      closeCreateDialog();
      openSnackbar("success", "Task created successfully.");
    } catch (error) {
      handleTaskMutationError(
        error,
        setCreateApiError,
        setCreateFieldErrors,
        "Unable to create the task right now. Please try again.",
        openSnackbar,
      );
    }
  }

  async function handleUpdateTask(values: TaskFormValues) {
    if (!editingTask) return;
    setEditApiError(null);
    setEditFieldErrors({});
    try {
      await updateTaskMutation.mutateAsync({
        taskId: editingTask.id,
        payload: normalizeTaskFormValues(values),
      });
      closeEditDialog();
      openSnackbar("success", "Task updated successfully.");
    } catch (error) {
      handleTaskMutationError(
        error,
        setEditApiError,
        setEditFieldErrors,
        "Unable to update the task right now. Please try again.",
        openSnackbar,
      );
    }
  }

  async function handleDeleteTask() {
    if (!deletingTask) return;
    setDeleteApiError(null);
    try {
      await deleteTaskMutation.mutateAsync({ taskId: deletingTask.id });
      closeDeleteDialog();
      openSnackbar("success", "Task deleted successfully.");
    } catch (error) {
      const message =
        error instanceof ApiError
          ? error.message
          : "Unable to delete the task right now. Please try again.";
      setDeleteApiError(message);
      openSnackbar("error", message);
    }
  }

  if (!projectId) {
    return <Alert severity="error">The project ID is missing from the route.</Alert>;
  }

  if (projectQuery.isLoading || statsQuery.isLoading) {
    return <ProjectDetailSkeleton />;
  }

  if (projectQuery.isError || statsQuery.isError) {
    return (
      <Alert severity="error">
        Unable to load this project right now. Please refresh and try again.
      </Alert>
    );
  }

  return (
    <>
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
          {statusSections.map((section) => (
            <Grid2 key={section.key} size={{ xs: 12, sm: 4 }}>
              <StatusSummaryCard
                label={section.label}
                icon={section.icon}
                total={stats?.status_counts[section.key] ?? 0}
              />
            </Grid2>
          ))}
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
                  <Select size="small" value={statusFilter} onChange={handleStatusChange}>
                    <MenuItem value="all">All statuses</MenuItem>
                    <MenuItem value="todo">Todo</MenuItem>
                    <MenuItem value="in_progress">In Progress</MenuItem>
                    <MenuItem value="done">Done</MenuItem>
                  </Select>

                  <Select size="small" value={assigneeFilter} onChange={handleAssigneeChange}>
                    <MenuItem value="all">All assignees</MenuItem>
                    {assigneeOptions.map((option) => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </Select>

                  <Button
                    variant="contained"
                    startIcon={<AddRoundedIcon />}
                    onClick={() => setIsCreateDialogOpen(true)}
                  >
                    Create task
                  </Button>
                </Stack>
              </Stack>

              <Divider />

              {project?.tasks.length === 0 ? (
                <EmptyState
                  title="No tasks yet"
                  description="This project does not have any tasks yet. Use the create task action to add the first item."
                />
              ) : filteredTasks.length === 0 ? (
                <EmptyState
                  title="No tasks match these filters"
                  description="Change the selected status or assignee filter to see more tasks."
                />
              ) : (
                <Grid2 container spacing={3}>
                  {statusSections
                    .filter((section) => statusFilter === "all" || section.key === statusFilter)
                    .map((section) => (
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
                              <StackHeader
                                label={section.label}
                                icon={section.icon}
                                count={tasksByStatus[section.key].length}
                              />

                              {tasksByStatus[section.key].length === 0 ? (
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
                                  {tasksByStatus[section.key].map((task) => (
                                    <TaskCard
                                      key={task.id}
                                      task={task}
                                      assigneeLabel={
                                        task.assignee_id
                                          ? assigneeNameMap.get(task.assignee_id) ??
                                            `User ${task.assignee_id.slice(0, 8)}...`
                                          : "Unassigned"
                                      }
                                      onEdit={() => {
                                        setEditApiError(null);
                                        setEditFieldErrors({});
                                        setEditingTask(task);
                                      }}
                                      onDelete={() => {
                                        setDeleteApiError(null);
                                        setDeletingTask(task);
                                      }}
                                    />
                                  ))}
                                </Stack>
                              )}
                            </Stack>
                          </CardContent>
                        </Card>
                      </Grid2>
                    ))}
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
              {(stats?.assignee_counts?.length ?? 0) > 0 ? (
                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                  {stats?.assignee_counts.map((assignee) => (
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

      <TaskDialog
        open={isCreateDialogOpen}
        title="Create task"
        submitLabel="Create task"
        initialValues={defaultTaskFormValues}
        assigneeOptions={assigneeOptions}
        apiError={createApiError}
        serverFieldErrors={createFieldErrors}
        isSubmitting={createTaskMutation.isPending}
        onClose={closeCreateDialog}
        onSubmit={handleCreateTask}
      />

      <TaskDialog
        open={Boolean(editingTask)}
        title="Edit task"
        submitLabel="Save changes"
        initialValues={editingTask ? mapTaskToFormValues(editingTask) : defaultTaskFormValues}
        assigneeOptions={assigneeOptions}
        apiError={editApiError}
        serverFieldErrors={editFieldErrors}
        isSubmitting={updateTaskMutation.isPending}
        onClose={closeEditDialog}
        onSubmit={handleUpdateTask}
      />

      <DeleteTaskDialog
        open={Boolean(deletingTask)}
        taskTitle={deletingTask?.title}
        apiError={deleteApiError}
        isSubmitting={deleteTaskMutation.isPending}
        onClose={closeDeleteDialog}
        onConfirm={handleDeleteTask}
      />

      <Snackbar
        open={snackbarState.open}
        autoHideDuration={4000}
        onClose={(_, reason) => {
          if (reason === "clickaway") return;
          setSnackbarState((current) => ({ ...current, open: false }));
        }}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
      >
        <MuiAlert severity={snackbarState.severity} variant="filled" elevation={6}>
          {snackbarState.message}
        </MuiAlert>
      </Snackbar>
    </>
  );
}

function ProjectDetailSkeleton() {
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

function StatusSummaryCard({
  label,
  icon: Icon,
  total,
}: {
  label: string;
  icon: typeof RadioButtonUncheckedRoundedIcon;
  total: number;
}) {
  return (
    <Card>
      <CardContent>
        <Stack spacing={1.5}>
          <Stack direction="row" spacing={1} alignItems="center">
            <Icon color="primary" fontSize="small" />
            <Typography color="text.secondary">{label}</Typography>
          </Stack>
          <Typography variant="h4">{String(total).padStart(2, "0")}</Typography>
          <Typography variant="body2" color="text.secondary">
            {label} tasks in the project
          </Typography>
        </Stack>
      </CardContent>
    </Card>
  );
}

function StackHeader({
  label,
  icon: Icon,
  count,
}: {
  label: string;
  icon: typeof RadioButtonUncheckedRoundedIcon;
  count: number;
}) {
  return (
    <Stack direction="row" spacing={1} justifyContent="space-between" alignItems="center">
      <Stack direction="row" spacing={1} alignItems="center">
        <Icon color="primary" fontSize="small" />
        <Typography variant="h6">{label}</Typography>
      </Stack>
      <Chip
        size="small"
        color="primary"
        label={`${count} task${count === 1 ? "" : "s"}`}
      />
    </Stack>
  );
}

function TaskCard({
  task,
  assigneeLabel,
  onEdit,
  onDelete,
}: {
  task: Task;
  assigneeLabel: string;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <Card variant="outlined" sx={{ borderRadius: 3 }}>
      <CardContent sx={{ p: 2 }}>
        <Stack spacing={1.5}>
          <Stack direction="row" spacing={1} justifyContent="space-between" alignItems="flex-start">
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
            {task.description?.trim() ? task.description : "No task description provided."}
          </Typography>

          <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
            <Chip size="small" variant="outlined" label={assigneeLabel} />
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

          <Stack direction="row" spacing={0.5} justifyContent="flex-end">
            <IconButton aria-label={`Edit ${task.title}`} color="primary" onClick={onEdit}>
              <EditOutlinedIcon fontSize="small" />
            </IconButton>
            <IconButton aria-label={`Delete ${task.title}`} color="error" onClick={onDelete}>
              <DeleteOutlineRoundedIcon fontSize="small" />
            </IconButton>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
}

function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <Stack spacing={1.5} alignItems="center" sx={{ py: 6, textAlign: "center" }}>
      <Typography variant="h6">{title}</Typography>
      <Typography color="text.secondary" maxWidth={560}>
        {description}
      </Typography>
    </Stack>
  );
}

function mapTaskToFormValues(task: Task): TaskFormValues {
  return {
    title: task.title,
    description: task.description ?? "",
    status: task.status,
    priority: task.priority,
    assignee_id: task.assignee_id ?? "unassigned",
    due_date: task.due_date ?? "",
  };
}

function normalizeTaskFormValues(values: TaskFormValues) {
  return {
    title: values.title.trim(),
    description: values.description?.trim() ? values.description.trim() : null,
    status: values.status,
    priority: values.priority,
    assignee_id:
      values.assignee_id && values.assignee_id !== "unassigned" ? values.assignee_id : null,
    due_date: values.due_date?.trim() ? values.due_date : null,
  };
}

function handleTaskMutationError(
  error: unknown,
  setApiError: (message: string) => void,
  setFieldErrors: (fields: Partial<Record<keyof TaskFormValues, string>>) => void,
  fallbackMessage: string,
  notify: (severity: "success" | "error", message: string) => void,
) {
  if (error instanceof ApiError) {
    setApiError(error.message);
    setFieldErrors(error.fields as Partial<Record<keyof TaskFormValues, string>>);
    notify("error", error.message);
    return;
  }

  setApiError(fallbackMessage);
  notify("error", fallbackMessage);
}

function priorityLabel(priority: Task["priority"]) {
  if (priority === "high") return "High priority";
  if (priority === "medium") return "Medium priority";
  return "Low priority";
}

function priorityColor(
  priority: Task["priority"],
): "default" | "success" | "warning" | "error" {
  if (priority === "high") return "error";
  if (priority === "medium") return "warning";
  return "success";
}
