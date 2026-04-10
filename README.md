# TaskFlow

## 1. Overview

TaskFlow is a full-stack task management application built for the take-home assignment. It supports user registration and login, project creation, task creation and updates, assignment to registered users, project-level authorization, project statistics, and real-time task updates on the project detail page.

The stack choices were intentionally conservative:

- Backend: Go with Chi, pgx, golang-migrate, bcrypt, JWT, and `slog`
- Frontend: React + TypeScript with React Router, TanStack Query, React Hook Form, Zod, and MUI
- Database: PostgreSQL
- Infrastructure: Docker and Docker Compose

Why these choices:

- Go keeps the API small, explicit, and easy to review. The standard library plus Chi was enough without introducing a heavy framework.
- Raw SQL through `pgx` keeps queries and authorization rules visible. For a take-home assignment, that is easier to defend in review than hiding behavior in an ORM.
- React + TypeScript gives a straightforward way to build a responsive UI with typed API contracts.
- TanStack Query handles server state, retries, cache invalidation, and optimistic task-status updates cleanly.
- PostgreSQL fits the relational model well: users, projects, tasks, foreign keys, and indexes are all first-class.
- Docker Compose makes the entire stack reproducible with one command.

This submission also includes SSE for real-time task updates. SSE was chosen over WebSockets because the application only needs one-way server-to-client updates for task events. That keeps the implementation smaller, easier to debug, and sufficient for the assignment’s realtime requirement without adding bidirectional connection complexity.

## 2. Architecture Decisions

The backend is split by domain into `auth`, `projects`, and `tasks`, each with handler/service/repository layers:

- Handlers own HTTP decoding, validation wiring, and status-code mapping.
- Services own business rules and authorization decisions.
- Repositories own SQL and database access.

This structure keeps the rules that matter most for the assignment visible:

- `GET /projects` returns projects the user owns or has assigned tasks in.
- `GET /projects/:id` allows access for the owner or a user assigned at least one task in that project.
- `PATCH /projects/:id` and `DELETE /projects/:id` are owner-only.
- `DELETE /tasks/:id` is allowed for the project owner or task creator only.

I added `creator_id` to tasks even though the assignment data model did not require it explicitly. That field exists to enforce the delete rule correctly and durably. Without `creator_id`, once a task is reassigned there is no reliable way to know who originally created it. Keeping it on the task preserves authorship and makes authorization deterministic.

For real-time updates, the backend publishes project-scoped task events through an in-memory SSE manager:

- `task_created`
- `task_updated`
- `task_deleted`

This is enough for a single-instance assignment deployment and integrates cleanly with the React project detail page.

The main tradeoffs were:

- There is no separate membership table. Project access is driven by ownership and task assignment only.
- `GET /users` is intentionally minimal and returns registered users for assignment; it is not a full user-management system.
- SSE is in-memory, so it is appropriate for one API instance, not a horizontally scaled multi-node deployment without an external pub/sub layer.
- Authentication uses JWT access tokens only. There is no refresh-token flow.

## 3. Running Locally

Prerequisites:

- Docker Desktop with Compose support

From the repository root:

```bash
cp .env.example .env
docker compose up --build
```

Services:

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- PostgreSQL: `localhost:5432`

Useful Docker Compose commands:

```bash
docker compose up --build
docker compose up -d
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs -f db
docker compose down
docker compose down -v
```

Notes:

- The backend applies migrations automatically on startup.
- The backend seeds predictable local data automatically on startup if the seed records are not already present.
- The frontend build is served by Nginx in Docker.
- The backend Dockerfile is multi-stage and the runtime image is distroless.

## 4. Running Migrations

No manual migration command is required for normal startup. Migrations run automatically when the backend container starts.

Standard startup:

```bash
cp .env.example .env
docker compose up --build
```

If you want to force a clean rebuild of the database and rerun migrations from scratch:

```bash
docker compose down -v
docker compose up --build
```

The migration files live in:

- `backend/migrations`

The local development seed file lives in:

- `backend/seeds/seed.sql`

## 5. Test Credentials

The seed data creates a test user automatically.

- Email: `test@example.com`
- Password: `password123`

The seed also creates:

- 1 project owned by the test user
- 3 tasks with `todo`, `in_progress`, and `done` statuses

## 6. API Reference

All non-auth endpoints require:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

### `POST /auth/register`

Request:

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "password123"
}
```

Response `201 Created`:

```json
{
  "token": "<jwt>",
  "user": {
    "id": "uuid",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "created_at": "2026-04-10T10:00:00Z"
  }
}
```

### `POST /auth/login`

Request:

```json
{
  "email": "jane@example.com",
  "password": "password123"
}
```

Response `200 OK`:

```json
{
  "token": "<jwt>",
  "user": {
    "id": "uuid",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "created_at": "2026-04-10T10:00:00Z"
  }
}
```

### `GET /health`

Response `200 OK`:

```json
{
  "status": "ok"
}
```

### `GET /users`

Returns registered users for task assignment.

Response `200 OK`:

```json
{
  "users": [
    {
      "id": "uuid",
      "name": "Jane Doe",
      "email": "jane@example.com"
    },
    {
      "id": "uuid",
      "name": "John Smith",
      "email": "john@example.com"
    }
  ]
}
```

### `GET /projects?page=1&limit=10`

Returns projects the current user owns or has assigned tasks in.

Response `200 OK`:

```json
{
  "projects": [
    {
      "id": "uuid",
      "name": "Website Redesign",
      "description": "Q2 project",
      "owner_id": "uuid",
      "created_at": "2026-04-10T10:00:00Z"
    }
  ],
  "page": 1,
  "limit": 10
}
```

### `POST /projects`

Request:

```json
{
  "name": "New Project",
  "description": "Optional description"
}
```

Response `201 Created`:

```json
{
  "id": "uuid",
  "name": "New Project",
  "description": "Optional description",
  "owner_id": "uuid",
  "created_at": "2026-04-10T10:00:00Z"
}
```

### `GET /projects/:id`

Accessible to the project owner or a user assigned at least one task in the project.

Response `200 OK`:

```json
{
  "id": "uuid",
  "name": "Website Redesign",
  "description": "Q2 project",
  "owner_id": "uuid",
  "created_at": "2026-04-10T10:00:00Z",
  "tasks": [
    {
      "id": "uuid",
      "title": "Design homepage",
      "description": "Create the first draft",
      "status": "in_progress",
      "priority": "high",
      "project_id": "uuid",
      "assignee_id": "uuid",
      "creator_id": "uuid",
      "due_date": "2026-04-15",
      "created_at": "2026-04-10T10:05:00Z",
      "updated_at": "2026-04-10T10:10:00Z"
    }
  ]
}
```

### `GET /projects/:id/stats`

Response `200 OK`:

```json
{
  "project_id": "uuid",
  "total_tasks": 3,
  "status_counts": {
    "todo": 1,
    "in_progress": 1,
    "done": 1
  },
  "assignee_counts": [
    {
      "assignee_id": "uuid",
      "assignee_name": "Jane Doe",
      "count": 2
    },
    {
      "assignee_id": null,
      "assignee_name": null,
      "count": 1
    }
  ]
}
```

### `GET /projects/:id/events`

SSE endpoint for live task updates. The frontend connects with the JWT as an `access_token` query parameter.

Example request:

```http
GET /projects/:id/events?access_token=<jwt>
Accept: text/event-stream
```

Example event payload:

```text
event: message
data: {"type":"task_updated","project_id":"uuid","task":{"id":"uuid","title":"Design homepage","status":"done","priority":"high","project_id":"uuid","assignee_id":"uuid","creator_id":"uuid","due_date":"2026-04-15","created_at":"2026-04-10T10:05:00Z","updated_at":"2026-04-10T10:20:00Z"}}
```

### `PATCH /projects/:id`

Owner only.

Request:

```json
{
  "name": "Updated Name",
  "description": "Updated description"
}
```

Response `200 OK`:

```json
{
  "id": "uuid",
  "name": "Updated Name",
  "description": "Updated description",
  "owner_id": "uuid",
  "created_at": "2026-04-10T10:00:00Z"
}
```

### `DELETE /projects/:id`

Owner only.

Response `200 OK`:

```json
{
  "message": "project deleted"
}
```

### `GET /projects/:id/tasks?status=todo&assignee=<user-id>&page=1&limit=20`

Response `200 OK`:

```json
{
  "tasks": [
    {
      "id": "uuid",
      "title": "Design homepage",
      "description": "Create the first draft",
      "status": "todo",
      "priority": "high",
      "project_id": "uuid",
      "assignee_id": "uuid",
      "creator_id": "uuid",
      "due_date": "2026-04-15",
      "created_at": "2026-04-10T10:05:00Z",
      "updated_at": "2026-04-10T10:10:00Z"
    }
  ],
  "page": 1,
  "limit": 20,
  "status": "todo",
  "assignee_id": "uuid"
}
```

### `POST /projects/:id/tasks`

Request:

```json
{
  "title": "Design homepage",
  "description": "Create the first draft",
  "assignee_id": "uuid",
  "due_date": "2026-04-15"
}
```

Response `201 Created`:

```json
{
  "id": "uuid",
  "title": "Design homepage",
  "description": "Create the first draft",
  "status": "todo",
  "priority": "medium",
  "project_id": "uuid",
  "assignee_id": "uuid",
  "creator_id": "uuid",
  "due_date": "2026-04-15",
  "created_at": "2026-04-10T10:05:00Z",
  "updated_at": "2026-04-10T10:05:00Z"
}
```

### `PATCH /tasks/:id`

Request:

```json
{
  "title": "Updated task title",
  "description": "Updated task description",
  "status": "done",
  "priority": "high",
  "assignee_id": "uuid",
  "due_date": "2026-04-20"
}
```

Response `200 OK`:

```json
{
  "id": "uuid",
  "title": "Updated task title",
  "description": "Updated task description",
  "status": "done",
  "priority": "high",
  "project_id": "uuid",
  "assignee_id": "uuid",
  "creator_id": "uuid",
  "due_date": "2026-04-20",
  "created_at": "2026-04-10T10:05:00Z",
  "updated_at": "2026-04-10T10:20:00Z"
}
```

### `DELETE /tasks/:id`

Allowed for the project owner or task creator only.

Response `200 OK`:

```json
{
  "message": "task deleted"
}
```

### Error examples

Validation error `400 Bad Request`:

```json
{
  "error": "validation failed",
  "fields": {
    "email": "email must be a valid email address"
  }
}
```

Unauthenticated `401 Unauthorized`:

```json
{
  "error": "unauthorized"
}
```

Forbidden `403 Forbidden`:

```json
{
  "error": "forbidden"
}
```

Not found `404 Not Found`:

```json
{
  "error": "project not found"
}
```

## 7. What You’d Do With More Time

- Add a proper project membership model instead of deriving collaboration solely from task assignment. That would make the access model more explicit and extensible.
- Replace the global `GET /users` endpoint with project-scoped assignable-user lookup or search if the user base grows.
- Persist SSE events through Redis, NATS, or PostgreSQL `LISTEN/NOTIFY` so realtime works across multiple API instances.
- Add refresh tokens, token revocation, and stronger session management.
- Add more backend tests around authorization edges, pagination, and invalid-input cases.
- Add total-count metadata to paginated endpoints so the frontend can render better paging controls.
- Improve project/task editing ergonomics for very large datasets, especially assignee search and task filtering.
- Remove the temporary local debugging route before production handoff.

The main shortcut was optimizing for a clean, reviewable assignment submission rather than building a general-purpose collaboration platform. The implementation focuses on correctness of the requested flows, explicit SQL and authorization rules, and a frontend that stays responsive and understandable without introducing unnecessary abstraction.
