import { Button, Card, CardContent, Stack, Typography } from "@mui/material";
import { Link as RouterLink } from "react-router-dom";

export function NotFoundPage() {
  return (
    <Card sx={{ maxWidth: 560, mx: "auto" }}>
      <CardContent sx={{ p: 4 }}>
        <Stack spacing={2}>
          <Typography variant="h4">Page not found</Typography>
          <Typography color="text.secondary">
            The route exists in the router shell, but this path is not defined.
          </Typography>
          <Button component={RouterLink} to="/projects" variant="contained" sx={{ width: "fit-content" }}>
            Back to projects
          </Button>
        </Stack>
      </CardContent>
    </Card>
  );
}
