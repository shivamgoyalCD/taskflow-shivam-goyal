# taskflow-shivam-goyal

Monorepo scaffold for the full-stack assignment.

Status: backend infrastructure scaffold is initialized. No business logic has been added yet.

## Structure

```text
taskflow-shivam-goyal/
  docker-compose.yml
  .env.example
  README.md
  backend/
  frontend/
```

## Backend

The backend now includes a Go API scaffold using Chi with:

- environment-based configuration loading
- JSON response helpers
- request ID, recovery, and logging middleware wiring
- `GET /health`

For local development, copy `.env.example` to `.env` before running the API.
