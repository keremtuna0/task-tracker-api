package task_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/keremtuna0/task-tracker-api/internal/database"
	"github.com/keremtuna0/task-tracker-api/internal/task"
)

func setupTestApp(t *testing.T) *fiber.App {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	migrationsDir := filepath.Join("..", "..", "migrations")
	if err := database.Migrate(db, migrationsDir); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	repo := task.NewSQLiteRepository(db)
	service := task.NewService(repo)
	handler := task.NewHandler(service)

	app := fiber.New()
	handler.Register(app)
	return app
}

func doRequest(t *testing.T, app *fiber.App, method, target string, body any) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, target, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s request: %v", method, target, err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	_ = resp.Body.Close()

	return resp, responseBody
}

func TestHandlerCRUDCycle(t *testing.T) {
	app := setupTestApp(t)

	createResp, createBody := doRequest(t, app, http.MethodPost, "/tasks", map[string]any{
		"title":       "Write tests",
		"description": "Integration coverage",
		"priority":    "high",
	})
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createResp.StatusCode, string(createBody))
	}

	var created map[string]any
	if err := json.Unmarshal(createBody, &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	id := int(created["id"].(float64))

	getResp, getBody := doRequest(t, app, http.MethodGet, "/tasks/1", nil)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", getResp.StatusCode, string(getBody))
	}

	listResp, listBody := doRequest(t, app, http.MethodGet, "/tasks?status=todo&priority=high", nil)
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listResp.StatusCode, string(listBody))
	}
	var listed []map[string]any
	if err := json.Unmarshal(listBody, &listed); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 task, got %d", len(listed))
	}

	updateResp, updateBody := doRequest(t, app, http.MethodPut, "/tasks/1", map[string]any{
		"status": "done",
	})
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", updateResp.StatusCode, string(updateBody))
	}

	deleteResp, _ := doRequest(t, app, http.MethodDelete, "/tasks/1", nil)
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", deleteResp.StatusCode)
	}

	notFoundResp, _ := doRequest(t, app, http.MethodGet, "/tasks/1", nil)
	if notFoundResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", notFoundResp.StatusCode)
	}

	updateDeletedResp, _ := doRequest(t, app, http.MethodPut, "/tasks/1", map[string]any{
		"title": "Should fail",
	})
	if updateDeletedResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 when updating deleted task, got %d", updateDeletedResp.StatusCode)
	}

	_ = id
}

func TestHandlerValidationAndSort(t *testing.T) {
	app := setupTestApp(t)

	badResp, badBody := doRequest(t, app, http.MethodPost, "/tasks", map[string]any{
		"title": "",
	})
	if badResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty title, got %d body=%s", badResp.StatusCode, string(badBody))
	}

	doRequest(t, app, http.MethodPost, "/tasks", map[string]any{
		"title":    "Earlier",
		"due_date": "2026-07-01T00:00:00Z",
	})
	doRequest(t, app, http.MethodPost, "/tasks", map[string]any{
		"title":    "Later",
		"due_date": "2026-07-10T00:00:00Z",
	})

	sortResp, sortBody := doRequest(t, app, http.MethodGet, "/tasks?sort=due_date&order=asc", nil)
	if sortResp.StatusCode != http.StatusOK {
		t.Fatalf("sort status = %d, body = %s", sortResp.StatusCode, string(sortBody))
	}

	var tasks []map[string]any
	if err := json.Unmarshal(sortBody, &tasks); err != nil {
		t.Fatalf("decode sort response: %v", err)
	}
	if len(tasks) < 2 {
		t.Fatalf("expected at least 2 tasks for sort test, got %d", len(tasks))
	}
	if tasks[0]["title"] != "Earlier" {
		t.Fatalf("expected Earlier first, got %v", tasks[0]["title"])
	}
}
