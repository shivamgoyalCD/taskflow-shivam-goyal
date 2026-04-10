import {
  Card,
  CardContent,
  Chip,
  Divider,
  Grid2,
  Stack,
  Typography,
} from "@mui/material";
import { useParams } from "react-router-dom";

const summaryCards = [
  { label: "Todo", value: "04" },
  { label: "In Progress", value: "02" },
  { label: "Done", value: "09" },
];

export function ProjectDetailsPage() {
  const { id } = useParams();

  return (
    <Stack spacing={4}>
      <Stack spacing={1.5}>
        <Chip
          label={`Project ID: ${id ?? "unknown"}`}
          color="secondary"
          variant="outlined"
          sx={{ width: "fit-content" }}
        />
        <Typography variant="h3">Project Overview</Typography>
        <Typography color="text.secondary" maxWidth={760}>
          Placeholder detail route for project summary, tasks, charts, and SSE-powered
          updates. The page is intentionally structured so data blocks can be replaced
          incrementally without reworking routing.
        </Typography>
      </Stack>

      <Grid2 container spacing={3}>
        {summaryCards.map((card) => (
          <Grid2 key={card.label} size={{ xs: 12, sm: 4 }}>
            <Card>
              <CardContent>
                <Stack spacing={1}>
                  <Typography color="text.secondary">{card.label}</Typography>
                  <Typography variant="h4">{card.value}</Typography>
                </Stack>
              </CardContent>
            </Card>
          </Grid2>
        ))}
      </Grid2>

      <Card>
        <CardContent sx={{ p: 3 }}>
          <Stack spacing={2}>
            <Typography variant="h5">Planned Sections</Typography>
            <Divider />
            <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
              <Chip label="Project stats" />
              <Chip label="Task list" />
              <Chip label="Filters" />
              <Chip label="Real-time stream" />
              <Chip label="Edit project drawer" />
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}
