import RadioButtonUncheckedRoundedIcon from "@mui/icons-material/RadioButtonUncheckedRounded";
import AutorenewRoundedIcon from "@mui/icons-material/AutorenewRounded";
import CheckCircleRoundedIcon from "@mui/icons-material/CheckCircleRounded";
import { useDroppable } from "@dnd-kit/core";
import { alpha } from "@mui/material/styles";
import { Box, Card, CardContent, Chip, Stack, Typography, useTheme } from "@mui/material";
import type { PropsWithChildren } from "react";
import type { Task } from "@/api/tasks";

type StatusColumnProps = PropsWithChildren<{
  status: Task["status"];
  count: number;
}>;

export function StatusColumn({ status, count, children }: StatusColumnProps) {
  const theme = useTheme();
  const { setNodeRef, isOver } = useDroppable({
    id: status,
  });

  const meta = statusMeta[status];

  return (
    <Card
      sx={{
        height: "100%",
        backgroundColor:
          theme.palette.mode === "dark" ? "rgba(15, 23, 42, 0.36)" : "rgba(255, 255, 255, 0.72)",
      }}
    >
      <CardContent sx={{ p: 2.5 }}>
        <Stack spacing={2}>
          <Stack direction="row" spacing={1} justifyContent="space-between" alignItems="center">
            <Stack direction="row" spacing={1} alignItems="center">
              <meta.icon color="primary" fontSize="small" />
              <Typography variant="h6">{meta.label}</Typography>
            </Stack>
            <Chip size="small" color="primary" label={`${count} task${count === 1 ? "" : "s"}`} />
          </Stack>

          <Box
            ref={setNodeRef}
            sx={{
              minHeight: 220,
              borderRadius: 3,
              border: `1px dashed ${isOver ? theme.palette.primary.main : theme.palette.divider}`,
              backgroundColor: isOver
                ? alpha(theme.palette.primary.main, theme.palette.mode === "dark" ? 0.2 : 0.08)
                : "transparent",
              transition: "background-color 160ms ease, border-color 160ms ease",
              p: 1,
            }}
          >
            {count > 0 ? (
              <Stack spacing={1.5}>{children}</Stack>
            ) : (
              <Box
                sx={{
                  minHeight: 180,
                  display: "grid",
                  placeItems: "center",
                  textAlign: "center",
                  px: 2,
                }}
              >
                <Typography variant="body2" color="text.secondary">
                  Drop a task here or adjust the filters to see more items.
                </Typography>
              </Box>
            )}
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}

const statusMeta = {
  todo: {
    label: "Todo",
    icon: RadioButtonUncheckedRoundedIcon,
  },
  in_progress: {
    label: "In Progress",
    icon: AutorenewRoundedIcon,
  },
  done: {
    label: "Done",
    icon: CheckCircleRoundedIcon,
  },
} satisfies Record<
  Task["status"],
  {
    label: string;
    icon: typeof RadioButtonUncheckedRoundedIcon;
  }
>;
