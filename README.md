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

## Integration Tests

The backend includes HTTP integration tests that exercise the real Chi router against an isolated PostgreSQL database created just for the test run.

To run them:

```powershell
docker compose up -d db
cd backend
go test ./cmd/api -v
```

Notes:

- The tests create and drop a temporary database automatically.
- By default the tests connect to `127.0.0.1:5432` with `postgres/postgres`.
- If your local setup is different, override these env vars before running tests:
  - `TEST_POSTGRES_HOST`
  - `TEST_POSTGRES_PORT`
  - `TEST_POSTGRES_USER`
  - `TEST_POSTGRES_PASSWORD`
  - `TEST_POSTGRES_ADMIN_DB`
