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
  useMediaQuery,
  useTheme,
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
  const theme = useTheme();
  const isSmallScreen = useMediaQuery(theme.breakpoints.down("sm"));

  return (
    <Dialog
      open={open}
      onClose={isSubmitting ? undefined : onClose}
      fullWidth
      maxWidth="xs"
      fullScreen={isSmallScreen}
      scroll="paper"
    >
      <DialogTitle>Delete task</DialogTitle>
      <DialogContent dividers sx={{ overflowX: "hidden" }}>
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
          disabled={isSubmitting}
          sx={{ width: { xs: "100%", sm: "auto" } }}
        >
          Cancel
        </Button>
        <Button
          color="error"
          variant="contained"
          disabled={isSubmitting}
          onClick={() => void onConfirm()}
          endIcon={isSubmitting ? <CircularProgress size={18} color="inherit" /> : null}
          sx={{ width: { xs: "100%", sm: "auto" } }}
        >
          Delete task
        </Button>
      </DialogActions>
    </Dialog>
  );
}
