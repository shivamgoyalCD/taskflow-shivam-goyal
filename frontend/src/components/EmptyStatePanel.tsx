import type { ReactNode } from "react";
import { Box, Stack, Typography } from "@mui/material";

type EmptyStatePanelProps = {
  icon: ReactNode;
  title: string;
  description: string;
  action?: ReactNode;
  compact?: boolean;
};

export function EmptyStatePanel({
  icon,
  title,
  description,
  action,
  compact = false,
}: EmptyStatePanelProps) {
  return (
    <Stack
      spacing={2}
      alignItems="center"
      textAlign="center"
      sx={{ py: compact ? 5 : 7, px: { xs: 2, sm: 4 } }}
    >
      <Box
        sx={{
          width: compact ? 56 : 64,
          height: compact ? 56 : 64,
          display: "grid",
          placeItems: "center",
          borderRadius: "50%",
          bgcolor: "action.hover",
          color: "primary.main",
        }}
      >
        {icon}
      </Box>
      <Stack spacing={1} maxWidth={560}>
        <Typography variant={compact ? "h6" : "h5"}>{title}</Typography>
        <Typography color="text.secondary">{description}</Typography>
      </Stack>
      {action ?? null}
    </Stack>
  );
}
