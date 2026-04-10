import { useDraggable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import DragIndicatorRoundedIcon from "@mui/icons-material/DragIndicatorRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import {
  Box,
  Card,
  CardContent,
  Chip,
  IconButton,
  Stack,
  Typography,
} from "@mui/material";
import type { Task } from "@/api/tasks";

type DraggableTaskCardProps = {
  task: Task;
  assigneeLabel: string;
  isStatusUpdating: boolean;
  onEdit: () => void;
  onDelete: () => void;
  overlay?: boolean;
};

export function DraggableTaskCard({
  task,
  assigneeLabel,
  isStatusUpdating,
  onEdit,
  onDelete,
  overlay = false,
}: DraggableTaskCardProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: overlay ? `${task.id}-overlay` : task.id,
    data: { task },
    disabled: overlay || isStatusUpdating,
  });

  return (
    <Card
      ref={setNodeRef}
      variant="outlined"
      sx={{
        borderRadius: 3,
        opacity: isDragging && !overlay ? 0.4 : 1,
        transform: transform ? CSS.Translate.toString(transform) : undefined,
        transition: transform ? "none" : "opacity 160ms ease, box-shadow 160ms ease",
        boxShadow: overlay ? 8 : undefined,
      }}
    >
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

          <Stack direction="row" spacing={0.5} justifyContent="space-between" alignItems="center">
            <IconButton
              size="small"
              aria-label={`Drag ${task.title}`}
              disabled={overlay || isStatusUpdating}
              sx={{ cursor: overlay || isStatusUpdating ? "default" : "grab", touchAction: "none" }}
              {...attributes}
              {...listeners}
            >
              <DragIndicatorRoundedIcon fontSize="small" />
            </IconButton>

            <Stack direction="row" spacing={0.5}>
              <IconButton
                aria-label={`Edit ${task.title}`}
                color="primary"
                disabled={overlay}
                onClick={onEdit}
              >
                <EditOutlinedIcon fontSize="small" />
              </IconButton>
              <IconButton
                aria-label={`Delete ${task.title}`}
                color="error"
                disabled={overlay}
                onClick={onDelete}
              >
                <DeleteOutlineRoundedIcon fontSize="small" />
              </IconButton>
            </Stack>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
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
