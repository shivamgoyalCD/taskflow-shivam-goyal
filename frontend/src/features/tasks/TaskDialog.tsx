import { useEffect } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  Stack,
  TextField,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import { Controller, useForm } from "react-hook-form";
import { taskFormSchema, type TaskFormValues } from "@/features/tasks/taskSchemas";

export type TaskAssigneeOption = {
  value: string;
  label: string;
};

type TaskDialogProps = {
  open: boolean;
  title: string;
  submitLabel: string;
  initialValues: TaskFormValues;
  assigneeOptions: TaskAssigneeOption[];
  apiError: string | null;
  serverFieldErrors?: Partial<Record<keyof TaskFormValues, string>>;
  isSubmitting: boolean;
  onClose: () => void;
  onSubmit: (values: TaskFormValues) => Promise<void>;
};

export function TaskDialog({
  open,
  title,
  submitLabel,
  initialValues,
  assigneeOptions,
  apiError,
  serverFieldErrors,
  isSubmitting,
  onClose,
  onSubmit,
}: TaskDialogProps) {
  const theme = useTheme();
  const isSmallScreen = useMediaQuery(theme.breakpoints.down("sm"));
  const {
    control,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting: isFormSubmitting },
  } = useForm<TaskFormValues>({
    resolver: zodResolver(taskFormSchema),
    defaultValues: initialValues,
  });

  useEffect(() => {
    if (open) {
      reset(initialValues);
    }
  }, [initialValues, open, reset]);

  useEffect(() => {
    if (!serverFieldErrors) {
      return;
    }

    for (const [fieldName, message] of Object.entries(serverFieldErrors)) {
      if (!message) {
        continue;
      }

      setError(fieldName as keyof TaskFormValues, {
        type: "server",
        message,
      });
    }
  }, [serverFieldErrors, setError]);

  const isBusy = isSubmitting || isFormSubmitting;

  return (
    <Dialog
      open={open}
      onClose={isBusy ? undefined : onClose}
      fullWidth
      maxWidth="sm"
      fullScreen={isSmallScreen}
      scroll="paper"
    >
      <DialogTitle>{title}</DialogTitle>

      <Stack component="form" onSubmit={handleSubmit(onSubmit)}>
        <DialogContent dividers sx={{ overflowX: "hidden" }}>
          <Stack spacing={2.5}>
            {apiError ? <Alert severity="error">{apiError}</Alert> : null}

            <Controller
              name="title"
              control={control}
              render={({ field }) => (
                <TextField
                  {...field}
                  label="Title"
                  autoFocus
                  fullWidth
                  error={Boolean(errors.title)}
                  helperText={errors.title?.message}
                />
              )}
            />

            <Controller
              name="description"
              control={control}
              render={({ field }) => (
                <TextField
                  {...field}
                  label="Description"
                  fullWidth
                  multiline
                  minRows={4}
                  error={Boolean(errors.description)}
                  helperText={errors.description?.message ?? "Optional"}
                />
              )}
            />

            <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
              <Controller
                name="status"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    select
                    fullWidth
                    label="Status"
                    error={Boolean(errors.status)}
                    helperText={errors.status?.message}
                  >
                    <MenuItem value="todo">Todo</MenuItem>
                    <MenuItem value="in_progress">In Progress</MenuItem>
                    <MenuItem value="done">Done</MenuItem>
                  </TextField>
                )}
              />

              <Controller
                name="priority"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    select
                    fullWidth
                    label="Priority"
                    error={Boolean(errors.priority)}
                    helperText={errors.priority?.message}
                  >
                    <MenuItem value="low">Low</MenuItem>
                    <MenuItem value="medium">Medium</MenuItem>
                    <MenuItem value="high">High</MenuItem>
                  </TextField>
                )}
              />
            </Stack>

            <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
              <Controller
                name="assignee_id"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    select
                    fullWidth
                    label="Assignee"
                    error={Boolean(errors.assignee_id)}
                    helperText={errors.assignee_id?.message ?? "Optional"}
                  >
                    {assigneeOptions.map((option) => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </TextField>
                )}
              />

              <Controller
                name="due_date"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    fullWidth
                    type="date"
                    label="Due date"
                    slotProps={{ inputLabel: { shrink: true } }}
                    error={Boolean(errors.due_date)}
                    helperText={errors.due_date?.message ?? "Optional"}
                  />
                )}
              />
            </Stack>
          </Stack>
        </DialogContent>

        <DialogActions
          sx={{
            px: 3,
            py: 2,
            flexDirection: { xs: "column-reverse", sm: "row" },
            alignItems: "stretch",
            gap: 1,
          }}
        >
          <Button
            onClick={onClose}
            color="inherit"
            disabled={isBusy}
            sx={{ width: { xs: "100%", sm: "auto" } }}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="contained"
            disabled={isBusy}
            endIcon={isBusy ? <CircularProgress size={18} color="inherit" /> : null}
            sx={{ width: { xs: "100%", sm: "auto" } }}
          >
            {submitLabel}
          </Button>
        </DialogActions>
      </Stack>
    </Dialog>
  );
}
