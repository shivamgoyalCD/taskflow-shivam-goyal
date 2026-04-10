import { useEffect, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
} from "@mui/material";
import { Controller, useForm } from "react-hook-form";
import { ApiError } from "@/api/client";
import {
  createProjectSchema,
  type CreateProjectFormValues,
} from "@/features/projects/projectSchemas";
import { useCreateProjectMutation } from "@/features/projects/useProjects";

type CreateProjectDialogProps = {
  open: boolean;
  onClose: () => void;
  onCreated?: () => void;
};

export function CreateProjectDialog({
  open,
  onClose,
  onCreated,
}: CreateProjectDialogProps) {
  const createProjectMutation = useCreateProjectMutation();
  const [apiError, setApiError] = useState<string | null>(null);

  const {
    control,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<CreateProjectFormValues>({
    resolver: zodResolver(createProjectSchema),
    defaultValues: {
      name: "",
      description: "",
    },
  });

  useEffect(() => {
    if (!open) {
      reset();
      setApiError(null);
      createProjectMutation.reset();
    }
  }, [createProjectMutation, open, reset]);

  async function onSubmit(values: CreateProjectFormValues) {
    setApiError(null);

    try {
      await createProjectMutation.mutateAsync({
        name: values.name.trim(),
        description: values.description?.trim() ? values.description.trim() : null,
      });
    } catch (error) {
      if (error instanceof ApiError) {
        setApiError(error.message);

        if (error.fields) {
          for (const [fieldName, message] of Object.entries(error.fields)) {
            setError(fieldName as keyof CreateProjectFormValues, {
              type: "server",
              message,
            });
          }
        }

        return;
      }

      setApiError("Unable to create the project right now. Please try again.");
      return;
    }

    onCreated?.();
    onClose();
  }

  const isBusy = isSubmitting || createProjectMutation.isPending;

  return (
    <Dialog open={open} onClose={isBusy ? undefined : onClose} fullWidth maxWidth="sm">
      <DialogTitle>Create project</DialogTitle>

      <Stack component="form" onSubmit={handleSubmit(onSubmit)}>
        <DialogContent dividers>
          <Stack spacing={2.5}>
            <Alert severity="info">
              Add a new project and it will appear in the paginated list after creation.
            </Alert>

            {apiError ? <Alert severity="error">{apiError}</Alert> : null}

            <Controller
              name="name"
              control={control}
              render={({ field }) => (
                <TextField
                  {...field}
                  label="Project name"
                  autoFocus
                  fullWidth
                  error={Boolean(errors.name)}
                  helperText={errors.name?.message}
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
          </Stack>
        </DialogContent>

        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button onClick={onClose} color="inherit" disabled={isBusy}>
            Cancel
          </Button>
          <Button
            type="submit"
            variant="contained"
            disabled={isBusy}
            endIcon={isBusy ? <CircularProgress size={18} color="inherit" /> : null}
          >
            Create project
          </Button>
        </DialogActions>
      </Stack>
    </Dialog>
  );
}
