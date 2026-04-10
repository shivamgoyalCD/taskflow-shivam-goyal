import HourglassTopRoundedIcon from "@mui/icons-material/HourglassTopRounded";
import { CircularProgress, Stack, Typography } from "@mui/material";

export function RouteLoadingState() {
  return (
    <Stack
      spacing={2}
      alignItems="center"
      justifyContent="center"
      sx={{ minHeight: "50vh", px: 2, textAlign: "center" }}
    >
      <HourglassTopRoundedIcon color="primary" />
      <CircularProgress />
      <Typography color="text.secondary">Loading the next view…</Typography>
    </Stack>
  );
}
