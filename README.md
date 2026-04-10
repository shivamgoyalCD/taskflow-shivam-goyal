# TaskFlow

A full-stack task management system with authentication, relational data, real-time updates, and a polished responsive UI.

Built by **Shivam Goyal** for the engineering take-home assignment.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture Decisions](#2-architecture-decisions)
3. [Running Locally](#3-running-locally)
4. [Running Migrations](#4-running-migrations)
5. [Test Credentials](#5-test-credentials)
6. [API Reference](#6-api-reference)
7. [What You'd Do With More Time](#7-what-youd-do-with-more-time)

---

## 1. Overview

TaskFlow lets users register, log in, create projects, add tasks to those projects, and assign tasks to themselves or other registered users. It includes project-level authorization, paginated listing, project statistics, drag-and-drop task management, dark mode, and real-time task updates via Server-Sent Events.

### Tech Stack

| Layer | Technology | Purpose |
|---|---|---|
| **Backend** | Go 1.24 | API server |
| **Router** | Chi v5 | HTTP routing and middleware |
| **Database** | PostgreSQL 16 | Persistent relational storage |
| **DB Driver** | pgx v5 | PostgreSQL driver with connection pooling |
| **Migrations** | golang-migrate v4 | Schema version control |
| **Auth** | bcrypt (cost 12) + JWT (HS256) | Password hashing and token-based auth |
| **Logging** | slog (JSON) | Structured logging |
| **Frontend** | React 18 + TypeScript | Single-page application |
| **UI Library** | MUI v6 (Material UI) | Component library and theming |
| **State** | TanStack Query v5 | Server state, caching, optimistic updates |
| **Forms** | React Hook Form + Zod | Form management with schema validation |
| **Drag & Drop** | dnd-kit | Kanban-style task board |
| **Routing** | React Router v6 | Client-side navigation |
| **Build** | Vite 5 | Frontend toolchain |
| **Containers** | Docker + Docker Compose | Full-stack orchestration |
| **Web Server** | Nginx 1.27 | Frontend static file serving |

### Why These Choices

- **Go** keeps the API small, explicit, and easy to review. Chi plus the standard library is sufficient without a heavy framework.
- **Raw SQL via pgx** keeps queries and authorization rules visible instead of hiding behavior behind an ORM. For a take-home, this is easier to defend in a code review.
- **React + TypeScript** gives typed API contracts and component composition. MUI provides accessible, responsive components out of the box.
- **TanStack Query** handles server state, retries, cache invalidation, and optimistic task-status updates cleanly without Redux or global state management.
- **SSE over WebSockets** because the application only needs one-way server-to-client task events. Simpler protocol, easier to debug, no bidirectional complexity.

---

## 2. Architecture Decisions

### System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Docker Compose                                   │
│                                                                             │
│  ┌──────────────┐      ┌──────────────────┐      ┌───────────────────┐     │
│  │              │      │                  │      │                   │     │
│  │   Frontend   │─────>│   Backend API    │─────>│   PostgreSQL 16   │     │
│  │   (Nginx)    │      │   (Go + Chi)     │      │                   │     │
│  │              │      │                  │      │   - users         │     │
│  │  Port 3000   │      │  Port 8080       │      │   - projects      │     │
│  │              │      │                  │      │   - tasks         │     │
│  │  React SPA   │<─ ─ ─│  REST + SSE     │      │                   │     │
│  │  (static)    │ SSE  │                  │      │   Port 5432       │     │
│  │              │      │                  │      │                   │     │
│  └──────────────┘      └──────────────────┘      └───────────────────┘     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Request flow**: Browser loads the React SPA from Nginx. All API calls go directly from the browser to the Go backend on port 8080. The backend authenticates requests via JWT, executes business logic, and queries PostgreSQL. For real-time updates, the frontend opens an SSE connection to the backend which pushes task events as they happen.

### Backend Architecture (Layered by Domain)

```
backend/
├── cmd/api/
│   ├── main.go                    # Entry point, wiring, server lifecycle
│   └── integration_test.go        # Integration test suite
│
├── internal/
│   ├── auth/                      # Authentication domain
│   │   ├── handler.go             #   HTTP handlers (register, login, list users)
│   │   ├── service.go             #   Business logic (hashing, token generation)
│   │   ├── repository.go          #   SQL queries (user CRUD)
│   │   └── jwt.go                 #   JWT token generation and parsing
│   │
│   ├── projects/                  # Projects domain
│   │   ├── handler.go             #   HTTP handlers (CRUD, stats, events)
│   │   ├── service.go             #   Business logic (authorization, access rules)
│   │   ├── repository.go          #   SQL queries (project + task queries)
│   │   └── events.go              #   SSE streaming endpoint
│   │
│   ├── tasks/                     # Tasks domain
│   │   ├── handler.go             #   HTTP handlers (CRUD with filters)
│   │   ├── service.go             #   Business logic (authorization, assignment)
│   │   └── repository.go          #   SQL queries (task CRUD with filters)
│   │
│   ├── middleware/                 # HTTP middleware
│   │   ├── auth.go                #   JWT authentication + context injection
│   │   ├── cors.go                #   CORS policy
│   │   ├── logging.go             #   Structured request logging
│   │   └── recovery.go            #   Panic recovery
│   │
│   ├── config/                    # Configuration
│   │   ├── config.go              #   Environment variable loading + validation
│   │   └── envfile.go             #   .env file parser
│   │
│   ├── db/                        # Database lifecycle
│   │   └── postgres.go            #   Connection, migrations, seeding
│   │
│   ├── realtime/                  # Real-time infrastructure
│   │   └── manager.go             #   In-memory pub/sub for SSE events
│   │
│   ├── response/                  # HTTP response helpers
│   │   └── json.go                #   Consistent JSON response formatting
│   │
│   └── validation/                # Input validation
│       ├── auth.go                #   Auth input rules
│       ├── projects.go            #   Project input + pagination rules
│       └── tasks.go               #   Task input rules
│
├── migrations/
│   ├── 000001_create_taskflow_schema.up.sql
│   └── 000001_create_taskflow_schema.down.sql
│
└── seeds/
    └── seed.sql
```

Each domain follows a **Handler → Service → Repository** pattern:

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────┐
│   HTTP       │     │  Validation  │     │   Service    │     │Repository│
│   Request    │────>│  (input      │────>│  (business   │────>│  (SQL    │
│              │     │   rules)     │     │   rules +    │     │  queries)│
│              │     │              │     │   authz)     │     │          │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────┘
                                                │
                                                │ on task mutation
                                                v
                                          ┌──────────────┐
                                          │  Realtime    │
                                          │  Manager     │
                                          │  (pub/sub)   │
                                          └──────┬───────┘
                                                 │ SSE
                                                 v
                                          ┌──────────────┐
                                          │  Connected   │
                                          │  Clients     │
                                          └──────────────┘
```

- **Handlers** own HTTP decoding, validation wiring, and status-code mapping.
- **Services** own business rules and authorization decisions.
- **Repositories** own SQL and database access.
- **Realtime Manager** receives events from services and fans them out to SSE subscribers.

### Frontend Architecture (Feature-Based)

```
frontend/src/
├── api/                           # API client layer
│   ├── client.ts                  #   Fetch wrapper with auth injection
│   ├── auth.ts                    #   Auth API calls
│   ├── projects.ts                #   Project API calls + types
│   ├── tasks.ts                   #   Task API calls + types
│   └── users.ts                   #   User API calls
│
├── app/                           # Application shell
│   ├── AppProviders.tsx           #   Provider composition (query, auth, theme)
│   ├── ThemeModeProvider.tsx      #   Dark/light mode with localStorage persistence
│   ├── queryClient.ts             #   TanStack Query configuration
│   ├── theme.ts                   #   MUI theme definitions
│   └── styles.css                 #   Global styles
│
├── components/                    # Shared components
│   ├── AppNavLink.tsx             #   Navigation link
│   ├── EmptyStatePanel.tsx        #   Empty state placeholder
│   └── RouteLoadingState.tsx      #   Route-level loading spinner
│
├── features/                      # Domain features
│   ├── auth/
│   │   ├── AuthContext.tsx         #   Auth state provider
│   │   ├── AuthFormCard.tsx        #   Reusable auth form
│   │   ├── ProtectedRoute.tsx      #   Route guard
│   │   ├── authSchemas.ts          #   Zod validation schemas
│   │   ├── authStorage.ts          #   localStorage session persistence
│   │   └── useAuthMutations.ts     #   Login/register mutations
│   │
│   ├── projects/
│   │   ├── CreateProjectDialog.tsx #   Project creation modal
│   │   ├── projectSchemas.ts       #   Zod schemas
│   │   ├── useProjectDetail.ts     #   Project detail query hooks
│   │   ├── useProjectEvents.ts     #   SSE connection hook
│   │   └── useProjects.ts          #   Projects list query hook
│   │
│   └── tasks/
│       ├── DeleteTaskDialog.tsx     #   Delete confirmation dialog
│       ├── DraggableTaskCard.tsx    #   Draggable task card for board
│       ├── StatusColumn.tsx         #   Kanban column component
│       ├── TaskBoard.tsx            #   Drag-and-drop board
│       ├── TaskDialog.tsx           #   Create/edit task modal
│       ├── taskSchemas.ts           #   Zod schemas
│       └── useTaskMutations.ts      #   Task CRUD + optimistic status mutations
│
├── layouts/
│   └── AppLayout.tsx              # App shell with navbar
│
├── pages/                         # Route pages
│   ├── LoginPage.tsx
│   ├── RegisterPage.tsx
│   ├── ProjectsPage.tsx
│   ├── ProjectDetailsPage.tsx
│   └── NotFoundPage.tsx
│
├── routes/
│   └── router.tsx                 # React Router configuration
│
├── types/
│   └── auth.ts                    # Auth type definitions
│
└── main.tsx                       # Entry point
```

### Database Schema (Entity-Relationship Diagram)

```
┌──────────────────────────┐
│          users            │
├──────────────────────────┤
│ id          UUID    [PK] │
│ name        TEXT    [NN] │
│ email       TEXT    [UQ] │
│ password    TEXT    [NN] │──── bcrypt hash, cost 12
│ created_at  TIMESTAMP    │
└──────────┬───────────────┘
           │
           │ 1
           │
     ┌─────┴──────┐
     │             │
     │ owner_id    │ creator_id    assignee_id
     │ [FK, NN]    │ [FK, NN]     [FK, nullable]
     │             │               │
     ▼             │               │
┌──────────────────┤───────┐       │
│      projects    │       │       │
├──────────────────┤───────┤       │
│ id          UUID │ [PK]  │       │
│ name        TEXT │ [NN]  │       │
│ description TEXT │ [null]│       │
│ owner_id    UUID │ [FK]──┘───>users.id (CASCADE)
│ created_at  TS   │       │       │
└────────┬─────────┘───────┘       │
         │                         │
         │ 1                       │
         │                         │
         ▼ N                       │
┌──────────────────────────┐       │
│          tasks            │       │
├──────────────────────────┤       │
│ id          UUID    [PK] │       │
│ title       TEXT    [NN] │       │
│ description TEXT    [null]│       │
│ status      TEXT    [NN] │──── CHECK: todo | in_progress | done
│ priority    TEXT    [NN] │──── CHECK: low | medium | high
│ project_id  UUID    [FK] │───>projects.id (CASCADE)
│ assignee_id UUID    [FK] │───>users.id (SET NULL) ◄──┘
│ creator_id  UUID    [FK] │───>users.id (CASCADE)
│ due_date    DATE   [null]│
│ created_at  TIMESTAMP    │
│ updated_at  TIMESTAMP    │
└──────────────────────────┘

Indexes:
  idx_projects_owner_id          ON projects(owner_id)
  idx_tasks_project_id           ON tasks(project_id)
  idx_tasks_assignee_id          ON tasks(assignee_id)
  idx_tasks_status               ON tasks(status)
  idx_tasks_project_id_status    ON tasks(project_id, status)
```

**Schema design notes:**

- `creator_id` was added to tasks even though the assignment data model did not require it explicitly. This field exists to enforce the delete authorization rule correctly: "project owner **or task creator** only." Without `creator_id`, once a task is reassigned, there is no reliable way to know who originally created it.
- `ON DELETE CASCADE` on `project_id` ensures deleting a project removes all its tasks atomically.
- `ON DELETE SET NULL` on `assignee_id` preserves tasks when a user is removed but nullifies the assignment.

### Authentication Flow

```
┌──────────┐                    ┌──────────────┐                 ┌──────────┐
│  Browser │                    │  Backend API │                 │PostgreSQL│
└────┬─────┘                    └──────┬───────┘                 └────┬─────┘
     │                                 │                              │
     │  POST /auth/register            │                              │
     │  {name, email, password}        │                              │
     │────────────────────────────────>│                              │
     │                                 │  Validate input              │
     │                                 │  Hash password (bcrypt 12)   │
     │                                 │  INSERT INTO users           │
     │                                 │─────────────────────────────>│
     │                                 │            user record       │
     │                                 │<─────────────────────────────│
     │                                 │  Generate JWT (24h expiry)   │
     │     {token, user}               │  Claims: user_id, email     │
     │<────────────────────────────────│                              │
     │                                 │                              │
     │  Store token in localStorage    │                              │
     │                                 │                              │
     │  GET /projects                  │                              │
     │  Authorization: Bearer <token>  │                              │
     │────────────────────────────────>│                              │
     │                                 │  Parse + validate JWT        │
     │                                 │  Extract user_id, email      │
     │                                 │  Inject into request context │
     │                                 │  Execute handler             │
     │                                 │─────────────────────────────>│
     │                                 │<─────────────────────────────│
     │         {projects: [...]}       │                              │
     │<────────────────────────────────│                              │
```

### Authorization Model

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Authorization Rules                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Project Visibility (GET /projects):                                    │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ User sees a project IF:                                         │    │
│  │   - They OWN the project          (owner_id = current_user)     │    │
│  │   - OR they have a TASK assigned   (assignee_id = current_user) │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Project Detail (GET /projects/:id):                                    │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ Same as project visibility. Owner OR assigned user.              │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Project Mutation (PATCH/DELETE /projects/:id):                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ OWNER ONLY. Returns 403 Forbidden for non-owners.               │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Task Update (PATCH /tasks/:id):                                        │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ Allowed for: Project owner, Task creator, OR Task assignee.     │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Task Delete (DELETE /tasks/:id):                                       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ Allowed for: Project owner OR Task creator ONLY.                │    │
│  │ (Assignee cannot delete.)                                       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Real-Time Event Flow (SSE)

```
┌──────────┐         ┌─────────────────────┐         ┌──────────┐
│ Client A │         │    Backend API       │         │ Client B │
│ (viewer) │         │                     │         │ (editor) │
└────┬─────┘         │  ┌───────────────┐  │         └────┬─────┘
     │                │  │   Realtime    │  │              │
     │  GET /projects │  │   Manager     │  │              │
     │  /:id/events   │  │               │  │              │
     │  ?access_token │  │  project_id → │  │              │
     │───────────────>│  │  [subscriber] │  │              │
     │                │  │               │  │              │
     │  : connected   │  └───────┬───────┘  │              │
     │<───────────────│          │           │              │
     │                │          │           │              │
     │                │          │           │  PATCH /tasks/:id
     │                │          │           │  {status: "done"}
     │                │          │           │<─────────────│
     │                │          │           │              │
     │                │   Update task in DB  │              │
     │                │   Publish event      │              │
     │                │          │           │  200 OK      │
     │                │          │           │─────────────>│
     │                │          │           │              │
     │  event:        │          │           │              │
     │  task_updated  │<─────────┘           │              │
     │  data: {...}   │                      │              │
     │<───────────────│                      │              │
     │                │                      │              │
     │  Update UI     │                      │              │
     │  (board +      │                      │              │
     │   stats)       │                      │              │
```

Events published: `task_created`, `task_updated`, `task_deleted`. Each event is scoped to a `project_id`. Only subscribers watching that project receive the event.

### Key Tradeoffs

| Decision | Rationale |
|---|---|
| No membership table | Project access is derived from ownership + task assignment. Simpler for the scope. A membership table would be the first addition for production. |
| `GET /users` returns all users | Intentionally minimal. Sufficient for task assignment in a small user base. Would become project-scoped search at scale. |
| In-memory SSE | Appropriate for single-instance deployment. Would need Redis/NATS pub/sub for horizontal scaling. |
| JWT-only auth (no refresh tokens) | Acceptable for 24h expiry in an assignment context. Production would add refresh tokens and revocation. |
| `creator_id` added to tasks | Not in the spec's data model, but necessary to correctly enforce "project owner or task creator" delete authorization without ambiguity after reassignment. |

### Bonus Features Implemented

| Feature | Implementation |
|---|---|
| Pagination | `?page=&limit=` on `GET /projects` and `GET /projects/:id/tasks` with validation and defaults |
| Project statistics | `GET /projects/:id/stats` returns task counts by status and by assignee |
| Integration tests | 6 test functions covering auth, task lifecycle, authorization, CORS, and SSE |
| Drag-and-drop | dnd-kit Kanban board with column-to-column task status changes |
| Dark mode | Persisted in localStorage, toggleable from the navbar |
| Real-time updates | SSE backend + EventSource frontend with live query cache patching |

---

## 3. Running Locally

**Prerequisites:** Docker Desktop with Compose support.

```bash
git clone https://github.com/shivam-goyal/taskflow-shivam-goyal
cd taskflow-shivam-goyal
cp .env.example .env
docker compose up --build
```

Once all three containers are healthy:

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| PostgreSQL | localhost:5432 |

### Useful Commands

```bash
# Start in background
docker compose up -d --build

# View logs
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs -f db

# Stop everything
docker compose down

# Full reset (wipes database volume)
docker compose down -v
docker compose up --build
```

### What Happens on Startup

1. PostgreSQL container starts and runs the health check.
2. Backend container waits for the database health check to pass.
3. Backend applies migrations automatically via `golang-migrate`.
4. Backend runs the seed script if seed data is not already present.
5. Frontend Nginx container starts and serves the React SPA.

The backend Dockerfile uses a **multi-stage build**: Go 1.24 Alpine as the build stage, Google distroless as the minimal runtime image.

---

## 4. Running Migrations

**No manual migration command is needed.** Migrations run automatically when the backend container starts.

To force a clean migration from scratch:

```bash
docker compose down -v
docker compose up --build
```

| Path | Contents |
|---|---|
| `backend/migrations/` | Up and down SQL migration files (golang-migrate format) |
| `backend/seeds/seed.sql` | Idempotent seed data for local development |

---

## 5. Test Credentials

The seed data creates a test user automatically on first startup:

```
Email:    test@example.com
Password: password123
```

The seed also creates:

- **1 project** owned by the test user
- **3 tasks** with `todo`, `in_progress`, and `done` statuses

---

## 6. API Reference

All non-auth endpoints require:

```
Authorization: Bearer <token>
Content-Type: application/json
```

### Authentication

#### `POST /auth/register`

Register a new user. Returns a JWT and the created user.

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

#### `POST /auth/login`

Authenticate an existing user. Returns a JWT.

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

### Users

#### `GET /users`

List all registered users (for task assignment dropdowns).

Response `200 OK`:

```json
{
  "users": [
    { "id": "uuid", "name": "Jane Doe", "email": "jane@example.com" },
    { "id": "uuid", "name": "John Smith", "email": "john@example.com" }
  ]
}
```

### Projects

#### `GET /projects?page=1&limit=10`

List projects the current user owns or has tasks assigned in. Supports pagination.

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

#### `POST /projects`

Create a new project. The authenticated user becomes the owner.

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

#### `GET /projects/:id`

Get project details including all tasks. Accessible to the project owner or any user with at least one assigned task in the project.

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

#### `PATCH /projects/:id`

Update project name or description. **Owner only.**

Request:

```json
{
  "name": "Updated Name",
  "description": "Updated description"
}
```

Response `200 OK`: Returns the updated project object.

#### `DELETE /projects/:id`

Delete the project and all its tasks. **Owner only.**

Response `204 No Content`

#### `GET /projects/:id/stats` (Bonus)

Get task counts grouped by status and by assignee.

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
    { "assignee_id": "uuid", "assignee_name": "Jane Doe", "count": 2 },
    { "assignee_id": null, "assignee_name": null, "count": 1 }
  ]
}
```

#### `GET /projects/:id/events` (Bonus -- SSE)

Server-Sent Events endpoint for live task updates. Authenticates via query parameter since EventSource does not support custom headers.

```
GET /projects/:id/events?access_token=<jwt>
Accept: text/event-stream
```

Event format:

```
event: task_updated
data: {"type":"task_updated","project_id":"uuid","task":{...}}
```

### Tasks

#### `GET /projects/:id/tasks?status=todo&assignee=<uuid>&page=1&limit=20`

List tasks in a project. Supports filtering by `status` and `assignee`, and pagination.

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

#### `POST /projects/:id/tasks`

Create a task in a project. Default status is `todo`, default priority is `medium`.

Request:

```json
{
  "title": "Design homepage",
  "description": "Create the first draft",
  "assignee_id": "uuid",
  "due_date": "2026-04-15"
}
```

Response `201 Created`: Returns the created task object.

#### `PATCH /tasks/:id`

Update any combination of task fields. Allowed for project owner, task creator, or task assignee.

Request (all fields optional):

```json
{
  "title": "Updated title",
  "description": "Updated description",
  "status": "done",
  "priority": "high",
  "assignee_id": "uuid",
  "due_date": "2026-04-20"
}
```

Response `200 OK`: Returns the updated task object.

#### `DELETE /tasks/:id`

Delete a task. **Project owner or task creator only.**

Response `204 No Content`

### Error Responses

All errors follow a consistent JSON structure:

| Status | Body | When |
|---|---|---|
| `400` | `{"error": "validation failed", "fields": {"email": "is required"}}` | Input validation failure |
| `401` | `{"error": "unauthorized"}` | Missing or invalid JWT |
| `403` | `{"error": "forbidden"}` | Valid JWT but insufficient permissions |
| `404` | `{"error": "not found"}` | Resource does not exist |
| `409` | `{"error": "email already exists"}` | Duplicate email on register |
| `500` | `{"error": "internal server error"}` | Unexpected server error |

---

## 7. What You'd Do With More Time

- **Project membership model.** Replace the implicit "access via task assignment" with an explicit membership table. This would make the authorization model more predictable and extensible (e.g., invite users to a project before assigning tasks).
- **Project-scoped user search.** Replace the global `GET /users` endpoint with a scoped lookup or autocomplete search, so the assignee dropdown scales beyond small teams.
- **Distributed SSE.** Persist events through Redis pub/sub, NATS, or PostgreSQL `LISTEN/NOTIFY` so real-time works across multiple horizontally scaled API instances.
- **Refresh tokens and revocation.** Add a refresh token flow, token rotation, and server-side revocation list for stronger session management.
- **More integration tests.** Cover authorization edge cases (e.g., outsider trying to create tasks in a project), pagination boundary conditions, and malformed input scenarios.
- **Total-count metadata.** Add `total_count` to paginated responses so the frontend can render accurate page counts and "Showing X of Y" indicators.
- **Frontend test coverage.** Add component tests with Vitest + Testing Library for auth flows, optimistic update rollbacks, and SSE reconnection behavior.

The main shortcut was optimizing for a clean, reviewable assignment submission rather than building a general-purpose collaboration platform. The implementation focuses on correctness of the requested flows, explicit SQL and authorization rules, and a frontend that stays responsive and understandable without introducing unnecessary abstraction.
