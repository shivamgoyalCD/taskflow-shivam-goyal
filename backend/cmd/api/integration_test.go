package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow-shivam-goyal/backend/internal/auth"
	"taskflow-shivam-goyal/backend/internal/config"
	"taskflow-shivam-goyal/backend/internal/db"
	"taskflow-shivam-goyal/backend/internal/projects"
	"taskflow-shivam-goyal/backend/internal/realtime"
	"taskflow-shivam-goyal/backend/internal/tasks"
)

var integrationSuite *testSuite

type testSuite struct {
	server     *httptest.Server
	pool       *pgxpool.Pool
	adminPool  *pgxpool.Pool
	database   string
	httpClient *http.Client
}

type authResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type projectResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	OwnerID     string  `json:"owner_id"`
}

type taskResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	ProjectID   string  `json:"project_id"`
	AssigneeID  *string `json:"assignee_id"`
	CreatorID   string  `json:"creator_id"`
	DueDate     *string `json:"due_date"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func TestMain(m *testing.M) {
	suite, err := newTestSuite()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up integration tests: %v\n", err)
		os.Exit(1)
	}

	integrationSuite = suite

	code := m.Run()

	if err := suite.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to tear down integration tests: %v\n", err)
		if code == 0 {
			code = 1
		}
	}

	os.Exit(code)
}

func TestIntegrationRegisterSuccess(t *testing.T) {
	suite := requireSuite(t)
	suite.ResetDatabase(t)

	payload := map[string]any{
		"name":     "Register Test User",
		"email":    uniqueEmail(t, "register"),
		"password": "password123",
	}

	response := suite.DoJSON(t, http.MethodPost, "/auth/register", "", payload)
	assertStatus(t, response, http.StatusCreated)

	var body authResponse
	decodeJSONResponse(t, response, &body)

	if body.Token == "" {
		t.Fatal("expected register response token to be set")
	}
	if body.User.ID == "" {
		t.Fatal("expected register response user id to be set")
	}
	if body.User.Email != strings.ToLower(payload["email"].(string)) {
		t.Fatalf("expected response email %q, got %q", strings.ToLower(payload["email"].(string)), body.User.Email)
	}

	var (
		storedName     string
		storedEmail    string
		storedPassword string
	)
	err := suite.pool.QueryRow(
		context.Background(),
		`SELECT name, email, password FROM users WHERE id = $1`,
		body.User.ID,
	).Scan(&storedName, &storedEmail, &storedPassword)
	if err != nil {
		t.Fatalf("query created user: %v", err)
	}

	if storedName != payload["name"] {
		t.Fatalf("expected stored name %q, got %q", payload["name"], storedName)
	}
	if storedEmail != strings.ToLower(payload["email"].(string)) {
		t.Fatalf("expected stored email %q, got %q", strings.ToLower(payload["email"].(string)), storedEmail)
	}
	if storedPassword == payload["password"] {
		t.Fatal("expected password to be hashed, but plaintext was stored")
	}
	if !strings.HasPrefix(storedPassword, "$2") {
		t.Fatalf("expected bcrypt hash prefix, got %q", storedPassword)
	}
}

func TestIntegrationLoginSuccess(t *testing.T) {
	suite := requireSuite(t)
	suite.ResetDatabase(t)

	email := uniqueEmail(t, "login-success")
	registerBody := map[string]any{
		"name":     "Login Success User",
		"email":    email,
		"password": "password123",
	}

	registerResponse := suite.DoJSON(t, http.MethodPost, "/auth/register", "", registerBody)
	assertStatus(t, registerResponse, http.StatusCreated)

	loginBody := map[string]any{
		"email":    email,
		"password": "password123",
	}

	loginResponse := suite.DoJSON(t, http.MethodPost, "/auth/login", "", loginBody)
	assertStatus(t, loginResponse, http.StatusOK)

	var body authResponse
	decodeJSONResponse(t, loginResponse, &body)

	if body.Token == "" {
		t.Fatal("expected login response token to be set")
	}
	if body.User.Email != strings.ToLower(email) {
		t.Fatalf("expected login response email %q, got %q", strings.ToLower(email), body.User.Email)
	}

	var userCount int
	err := suite.pool.QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM users WHERE email = $1`,
		strings.ToLower(email),
	).Scan(&userCount)
	if err != nil {
		t.Fatalf("count users by email: %v", err)
	}
	if userCount != 1 {
		t.Fatalf("expected 1 matching user row, got %d", userCount)
	}
}

func TestIntegrationLoginFailure(t *testing.T) {
	suite := requireSuite(t)
	suite.ResetDatabase(t)

	email := uniqueEmail(t, "login-failure")
	registerBody := map[string]any{
		"name":     "Login Failure User",
		"email":    email,
		"password": "password123",
	}

	registerResponse := suite.DoJSON(t, http.MethodPost, "/auth/register", "", registerBody)
	assertStatus(t, registerResponse, http.StatusCreated)

	loginBody := map[string]any{
		"email":    email,
		"password": "wrong-password",
	}

	loginResponse := suite.DoJSON(t, http.MethodPost, "/auth/login", "", loginBody)
	assertStatus(t, loginResponse, http.StatusUnauthorized)

	var body errorResponse
	decodeJSONResponse(t, loginResponse, &body)

	if body.Error != "invalid credentials" {
		t.Fatalf("expected invalid credentials error, got %q", body.Error)
	}
}

func TestIntegrationCreateAndUpdateTaskFlow(t *testing.T) {
	suite := requireSuite(t)
	suite.ResetDatabase(t)

	registerBody := map[string]any{
		"name":     "Task Flow User",
		"email":    uniqueEmail(t, "task-flow"),
		"password": "password123",
	}

	registerResponse := suite.DoJSON(t, http.MethodPost, "/auth/register", "", registerBody)
	assertStatus(t, registerResponse, http.StatusCreated)

	var register authResponse
	decodeJSONResponse(t, registerResponse, &register)

	createProjectBody := map[string]any{
		"name":        "Integration Project",
		"description": "Task flow integration test",
	}

	projectResp := suite.DoJSON(t, http.MethodPost, "/projects", register.Token, createProjectBody)
	assertStatus(t, projectResp, http.StatusCreated)

	var project projectResponse
	decodeJSONResponse(t, projectResp, &project)

	createTaskBody := map[string]any{
		"title":       "Initial Task Title",
		"description": "Create and update task flow",
		"due_date":    "2026-04-20",
	}

	taskCreateResponse := suite.DoJSON(
		t,
		http.MethodPost,
		"/projects/"+project.ID+"/tasks",
		register.Token,
		createTaskBody,
	)
	assertStatus(t, taskCreateResponse, http.StatusCreated)

	var createdTask taskResponse
	decodeJSONResponse(t, taskCreateResponse, &createdTask)

	if createdTask.Status != "todo" {
		t.Fatalf("expected default task status todo, got %q", createdTask.Status)
	}
	if createdTask.Priority != "medium" {
		t.Fatalf("expected default task priority medium, got %q", createdTask.Priority)
	}
	if createdTask.CreatorID != register.User.ID {
		t.Fatalf("expected creator id %q, got %q", register.User.ID, createdTask.CreatorID)
	}

	updateTaskBody := map[string]any{
		"title":       "Updated Task Title",
		"description": "Updated task description",
		"status":      "done",
		"priority":    "high",
		"due_date":    "2026-04-30",
	}

	taskUpdateResponse := suite.DoJSON(
		t,
		http.MethodPatch,
		"/tasks/"+createdTask.ID,
		register.Token,
		updateTaskBody,
	)
	assertStatus(t, taskUpdateResponse, http.StatusOK)

	var updatedTask taskResponse
	decodeJSONResponse(t, taskUpdateResponse, &updatedTask)

	if updatedTask.Status != "done" {
		t.Fatalf("expected updated task status done, got %q", updatedTask.Status)
	}
	if updatedTask.Priority != "high" {
		t.Fatalf("expected updated task priority high, got %q", updatedTask.Priority)
	}
	if updatedTask.Title != "Updated Task Title" {
		t.Fatalf("expected updated task title to persist, got %q", updatedTask.Title)
	}

	var (
		storedTitle       string
		storedDescription *string
		storedStatus      string
		storedPriority    string
		storedCreatorID   string
		storedProjectID   string
		storedDueDate     *string
	)
	err := suite.pool.QueryRow(
		context.Background(),
		`
			SELECT
				title,
				description,
				status,
				priority,
				creator_id,
				project_id,
				CASE
					WHEN due_date IS NULL THEN NULL
					ELSE TO_CHAR(due_date, 'YYYY-MM-DD')
				END AS due_date
			FROM tasks
			WHERE id = $1
		`,
		createdTask.ID,
	).Scan(
		&storedTitle,
		&storedDescription,
		&storedStatus,
		&storedPriority,
		&storedCreatorID,
		&storedProjectID,
		&storedDueDate,
	)
	if err != nil {
		t.Fatalf("query updated task: %v", err)
	}

	if storedTitle != "Updated Task Title" {
		t.Fatalf("expected stored task title %q, got %q", "Updated Task Title", storedTitle)
	}
	if storedDescription == nil || *storedDescription != "Updated task description" {
		t.Fatalf("expected stored description to be updated, got %#v", storedDescription)
	}
	if storedStatus != "done" {
		t.Fatalf("expected stored task status done, got %q", storedStatus)
	}
	if storedPriority != "high" {
		t.Fatalf("expected stored task priority high, got %q", storedPriority)
	}
	if storedCreatorID != register.User.ID {
		t.Fatalf("expected stored creator id %q, got %q", register.User.ID, storedCreatorID)
	}
	if storedProjectID != project.ID {
		t.Fatalf("expected stored project id %q, got %q", project.ID, storedProjectID)
	}
	if storedDueDate == nil || *storedDueDate != "2026-04-30" {
		t.Fatalf("expected stored due date to be updated, got %#v", storedDueDate)
	}
}

func TestIntegrationRegisterPreflightCORS(t *testing.T) {
	suite := requireSuite(t)
	suite.ResetDatabase(t)

	request, err := http.NewRequest(http.MethodOptions, suite.server.URL+"/auth/register", nil)
	if err != nil {
		t.Fatalf("create CORS preflight request: %v", err)
	}

	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type")

	response, err := suite.httpClient.Do(request)
	if err != nil {
		t.Fatalf("execute CORS preflight request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected preflight status %d, got %d", http.StatusNoContent, response.StatusCode)
	}

	if got := response.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected Access-Control-Allow-Origin header to echo dev origin, got %q", got)
	}
	if got := response.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodPost) {
		t.Fatalf("expected Access-Control-Allow-Methods to include POST, got %q", got)
	}
	if got := response.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(strings.ToLower(got), "content-type") {
		t.Fatalf("expected Access-Control-Allow-Headers to include Content-Type, got %q", got)
	}
}

func newTestSuite() (*testSuite, error) {
	baseConfig, err := integrationPostgresConfig()
	if err != nil {
		return nil, err
	}

	adminConfig := baseConfig
	adminConfig.Database = lookupStringEnv("TEST_POSTGRES_ADMIN_DB", "postgres")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	adminPool, err := db.Open(ctx, adminConfig)
	if err != nil {
		return nil, fmt.Errorf("open admin postgres pool: %w", err)
	}

	databaseName := fmt.Sprintf("taskflow_integration_%d", time.Now().UnixNano())

	if _, err := adminPool.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, databaseName)); err != nil {
		adminPool.Close()
		return nil, fmt.Errorf("create test database %q: %w", databaseName, err)
	}

	testConfig := baseConfig
	testConfig.Database = databaseName

	pool, err := db.Open(ctx, testConfig)
	if err != nil {
		dropDatabase(context.Background(), adminPool, databaseName)
		adminPool.Close()
		return nil, fmt.Errorf("open test postgres pool: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	migrationCtx, cancelMigration := db.StartupContext(context.Background())
	defer cancelMigration()

	if err := db.RunMigrations(migrationCtx, logger, testConfig, pool); err != nil {
		pool.Close()
		dropDatabase(context.Background(), adminPool, databaseName)
		adminPool.Close()
		return nil, fmt.Errorf("run test migrations: %w", err)
	}

	jwtSecret := lookupStringEnv("JWT_SECRET", "change-me")
	jwtExpiryHours := lookupIntEnv("JWT_EXPIRY_HOURS", 24)
	jwtExpiry := time.Duration(jwtExpiryHours) * time.Hour

	realtimeManager := realtime.NewManager()
	authRepository := auth.NewRepository(pool)
	authService := auth.NewService(authRepository, auth.NewJWTManager(jwtSecret, jwtExpiry))
	authHandler := auth.NewHandler(logger, authService)

	projectsRepository := projects.NewRepository(pool)
	projectsService := projects.NewService(projectsRepository)
	projectsHandler := projects.NewHandler(logger, projectsService, realtimeManager)

	tasksRepository := tasks.NewRepository(pool)
	tasksService := tasks.NewService(tasksRepository, realtimeManager)
	tasksHandler := tasks.NewHandler(logger, tasksService)

	app := &application{
		logger:          logger,
		db:              pool,
		authHandler:     authHandler,
		projectsHandler: projectsHandler,
		tasksHandler:    tasksHandler,
		jwtManager:      auth.NewJWTManager(jwtSecret, jwtExpiry),
	}

	server := httptest.NewServer(newRouter(app))

	return &testSuite{
		server:     server,
		pool:       pool,
		adminPool:  adminPool,
		database:   databaseName,
		httpClient: server.Client(),
	}, nil
}

func (s *testSuite) Close() error {
	if s.server != nil {
		s.server.Close()
	}
	if s.pool != nil {
		s.pool.Close()
	}
	if s.adminPool != nil {
		if err := dropDatabase(context.Background(), s.adminPool, s.database); err != nil {
			s.adminPool.Close()
			return err
		}
		s.adminPool.Close()
	}

	return nil
}

func (s *testSuite) ResetDatabase(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.pool.Exec(ctx, `TRUNCATE TABLE tasks, projects, users CASCADE`)
	if err != nil {
		t.Fatalf("truncate test tables: %v", err)
	}
}

func (s *testSuite) DoJSON(t *testing.T, method string, path string, token string, body any) *http.Response {
	t.Helper()

	var requestBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		requestBody = bytes.NewReader(payload)
	}

	request, err := http.NewRequest(method, s.server.URL+path, requestBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := s.httpClient.Do(request)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}

	t.Cleanup(func() {
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
	})

	return response
}

func requireSuite(t *testing.T) *testSuite {
	t.Helper()

	if integrationSuite == nil {
		t.Fatal("integration suite is not initialized")
	}

	return integrationSuite
}

func assertStatus(t *testing.T, response *http.Response, expected int) {
	t.Helper()

	if response.StatusCode == expected {
		return
	}

	body, _ := io.ReadAll(response.Body)
	t.Fatalf("expected status %d, got %d: %s", expected, response.StatusCode, string(body))
}

func decodeJSONResponse(t *testing.T, response *http.Response, destination any) {
	t.Helper()

	if contentType := response.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected application/json content type, got %q", contentType)
	}

	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		t.Fatalf("decode JSON response: %v", err)
	}
}

func uniqueEmail(t *testing.T, prefix string) string {
	t.Helper()

	return fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano())
}

func integrationPostgresConfig() (config.PostgresConfig, error) {
	port, err := lookupPortEnv("TEST_POSTGRES_PORT", "POSTGRES_PORT", 5432)
	if err != nil {
		return config.PostgresConfig{}, err
	}

	return config.PostgresConfig{
		Host:     lookupStringEnvWithFallback("TEST_POSTGRES_HOST", "POSTGRES_HOST", "127.0.0.1"),
		Port:     port,
		Database: "",
		User:     lookupStringEnvWithFallback("TEST_POSTGRES_USER", "POSTGRES_USER", "postgres"),
		Password: lookupStringEnvWithFallback("TEST_POSTGRES_PASSWORD", "POSTGRES_PASSWORD", "postgres"),
	}, nil
}

func lookupPortEnv(primary string, fallback string, defaultValue int) (int, error) {
	value := lookupStringEnvWithFallback(primary, fallback, strconv.Itoa(defaultValue))
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid postgres port %q", value)
	}

	return port, nil
}

func lookupStringEnvWithFallback(primary string, fallback string, defaultValue string) string {
	if value := lookupStringEnv(primary, ""); value != "" {
		return value
	}

	if value := lookupStringEnv(fallback, ""); value != "" {
		return value
	}

	return defaultValue
}

func lookupStringEnv(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	return value
}

func lookupIntEnv(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func dropDatabase(ctx context.Context, adminPool *pgxpool.Pool, databaseName string) error {
	if adminPool == nil || databaseName == "" {
		return nil
	}

	dropCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if _, err := adminPool.Exec(dropCtx, `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`, databaseName); err != nil {
		return fmt.Errorf("terminate test database connections: %w", err)
	}

	if _, err := adminPool.Exec(dropCtx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, databaseName)); err != nil {
		return fmt.Errorf("drop test database %q: %w", databaseName, err)
	}

	return nil
}
