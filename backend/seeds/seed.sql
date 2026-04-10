-- Local development seed data for the Taskflow schema.
-- Apply this after running the initial migrations.
-- Uses fixed UUIDs and explicit timestamps so local test data is predictable.
-- The user password below is stored as a bcrypt hash, not as plaintext.
-- Inserts are idempotent so this file is safe to execute multiple times.

BEGIN;

-- Seed user
INSERT INTO users (
    id,
    name,
    email,
    password,
    created_at
) VALUES (
    '11111111-1111-4111-8111-111111111111',
    'Test User',
    'test@example.com',
    '$2a$12$ESSGnHrP0KolDLnep6AMs.cYzGQ63XgNyiZR2KnKrXPHA1seCTjAK',
    '2026-01-10 09:00:00'
) ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    email = EXCLUDED.email,
    password = EXCLUDED.password,
    created_at = EXCLUDED.created_at;

-- Seed project owned by the test user
INSERT INTO projects (
    id,
    name,
    description,
    owner_id,
    created_at
) VALUES (
    '22222222-2222-4222-8222-222222222222',
    'Taskflow Demo Project',
    'Predictable local project seed for development and API testing.',
    '11111111-1111-4111-8111-111111111111',
    '2026-01-10 09:05:00'
) ON CONFLICT DO NOTHING;

-- Seed tasks covering each supported status value
INSERT INTO tasks (
    id,
    title,
    description,
    status,
    priority,
    project_id,
    assignee_id,
    creator_id,
    due_date,
    created_at,
    updated_at
) VALUES
(
    '33333333-3333-4333-8333-333333333331',
    'Draft project scope',
    'Create the initial scope outline for the seeded demo project.',
    'todo',
    'high',
    '22222222-2222-4222-8222-222222222222',
    '11111111-1111-4111-8111-111111111111',
    '11111111-1111-4111-8111-111111111111',
    '2026-01-15',
    '2026-01-10 09:10:00',
    '2026-01-10 09:10:00'
),
(
    '33333333-3333-4333-8333-333333333332',
    'Build API scaffold',
    'Implement the core backend scaffold and wiring for the project.',
    'in_progress',
    'medium',
    '22222222-2222-4222-8222-222222222222',
    '11111111-1111-4111-8111-111111111111',
    '11111111-1111-4111-8111-111111111111',
    '2026-01-18',
    '2026-01-10 09:20:00',
    '2026-01-11 11:00:00'
),
(
    '33333333-3333-4333-8333-333333333333',
    'Create database schema',
    'Set up the initial PostgreSQL tables, constraints, and indexes.',
    'done',
    'low',
    '22222222-2222-4222-8222-222222222222',
    '11111111-1111-4111-8111-111111111111',
    '11111111-1111-4111-8111-111111111111',
    '2026-01-12',
    '2026-01-10 09:30:00',
    '2026-01-12 17:30:00'
) ON CONFLICT DO NOTHING;

COMMIT;
