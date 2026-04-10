import {
  Alert,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
} from "@mui/material";

type DeleteTaskDialogProps = {
  open: boolean;
  taskTitle?: string;
  apiError: string | null;
  isSubmitting: boolean;
  onClose: () => void;
  onConfirm: () => Promise<void>;
};

export function DeleteTaskDialog({
  open,
  taskTitle,
  apiError,
  isSubmitting,
  onClose,
  onConfirm,
}: DeleteTaskDialogProps) {
  return (
    <Dialog open={open} onClose={isSubmitting ? undefined : onClose} fullWidth maxWidth="xs">
      <DialogTitle>Delete task</DialogTitle>
      <DialogContent dividers>
        <Stack spacing={2}>
          <Alert severity="warning">
            This action cannot be undone. The task will be removed from the project.
          </Alert>

          {apiError ? <Alert severity="error">{apiError}</Alert> : null}

          <Typography color="text.secondary">
            {taskTitle
              ? `Delete "${taskTitle}" from this project?`
              : "Delete this task from the project?"}
          </Typography>
        </Stack>
      </DialogContent>
      <DialogActions sx={{ px: 3, py: 2 }}>
        <Button onClick={onClose} color="inherit" disabled={isSubmitting}>
          Cancel
        </Button>
        <Button
          color="error"
          variant="contained"
          disabled={isSubmitting}
          onClick={() => void onConfirm()}
          endIcon={isSubmitting ? <CircularProgress size={18} color="inherit" /> : null}
        >
          Delete task
        </Button>
      </DialogActions>
    </Dialog>
  );
}
