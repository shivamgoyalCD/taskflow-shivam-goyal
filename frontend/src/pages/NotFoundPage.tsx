import SearchOffRoundedIcon from "@mui/icons-material/SearchOffRounded";
import { Button, Card, CardContent } from "@mui/material";
import { Link as RouterLink } from "react-router-dom";
import { EmptyStatePanel } from "@/components/EmptyStatePanel";

export function NotFoundPage() {
  return (
    <Card sx={{ maxWidth: 560, mx: "auto" }}>
      <CardContent sx={{ p: 4 }}>
        <EmptyStatePanel
          icon={<SearchOffRoundedIcon />}
          title="Page not found"
          description="The page you requested does not exist or is no longer available."
          action={
            <Button component={RouterLink} to="/projects" variant="contained">
              Back to projects
            </Button>
          }
        />
      </CardContent>
    </Card>
  );
}
