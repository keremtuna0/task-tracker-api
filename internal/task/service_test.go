package task

import (
	"context"
	"testing"
)

type mockRepository struct {
	tasks map[int64]*Task
	next  int64
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		tasks: make(map[int64]*Task),
		next:  1,
	}
}

func (m *mockRepository) Create(_ context.Context, task *Task) error {
	task.ID = m.next
	m.next++
	copy := *task
	m.tasks[task.ID] = &copy
	return nil
}

func (m *mockRepository) List(_ context.Context, filter ListFilter) ([]Task, error) {
	result := make([]Task, 0)
	for _, t := range m.tasks {
		if t.DeletedAt != nil {
			continue
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.Priority != "" && t.Priority != filter.Priority {
			continue
		}
		result = append(result, *t)
	}
	return result, nil
}

func (m *mockRepository) GetByID(_ context.Context, id int64) (*Task, error) {
	task, ok := m.tasks[id]
	if !ok || task.DeletedAt != nil {
		return nil, ErrNotFound
	}
	copy := *task
	return &copy, nil
}

func (m *mockRepository) Update(_ context.Context, task *Task) error {
	existing, ok := m.tasks[task.ID]
	if !ok || existing.DeletedAt != nil {
		return ErrNotFound
	}
	copy := *task
	m.tasks[task.ID] = &copy
	return nil
}

func (m *mockRepository) SoftDelete(_ context.Context, id int64, deletedAt string) error {
	task, ok := m.tasks[id]
	if !ok || task.DeletedAt != nil {
		return ErrNotFound
	}
	task.DeletedAt = &deletedAt
	task.UpdatedAt = deletedAt
	return nil
}

func TestServiceCreateValidation(t *testing.T) {
	service := NewService(newMockRepository())

	_, err := service.Create(context.Background(), CreateInput{Title: "   "})
	if err == nil {
		t.Fatal("expected validation error for empty title")
	}
	if !IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}

	_, err = service.Create(context.Background(), CreateInput{
		Title:  "Task",
		Status: "invalid",
	})
	if err == nil || !IsValidationError(err) {
		t.Fatalf("expected invalid status error, got %v", err)
	}

	task, err := service.Create(context.Background(), CreateInput{Title: "Learn Go"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.Status != StatusTodo {
		t.Fatalf("expected default status todo, got %s", task.Status)
	}
	if task.Priority != PriorityMedium {
		t.Fatalf("expected default priority medium, got %s", task.Priority)
	}
}

func TestServiceUpdateNotFoundAndValidation(t *testing.T) {
	service := NewService(newMockRepository())

	_, err := service.Update(context.Background(), 99, UpdateInput{})
	if err == nil || !IsValidationError(err) && err != ErrNotFound {
		if err != ErrNotFound {
			t.Fatalf("expected not found, got %v", err)
		}
	}

	created, err := service.Create(context.Background(), CreateInput{Title: "Task"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	empty := "   "
	_, err = service.Update(context.Background(), created.ID, UpdateInput{Title: &empty})
	if err == nil || !IsValidationError(err) {
		t.Fatalf("expected empty title validation error, got %v", err)
	}
}

func TestServiceSoftDeleteGuards(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo)

	if err := service.Delete(context.Background(), 1); err != ErrNotFound {
		t.Fatalf("expected not found for missing task, got %v", err)
	}

	created, err := service.Create(context.Background(), CreateInput{Title: "Task"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	if err := service.Delete(context.Background(), created.ID); err != ErrNotFound {
		t.Fatalf("expected not found for already deleted task, got %v", err)
	}

	_, err = service.GetByID(context.Background(), created.ID)
	if err != ErrNotFound {
		t.Fatalf("expected not found after soft delete, got %v", err)
	}

	_, err = service.Update(context.Background(), created.ID, UpdateInput{})
	if err != ErrNotFound {
		t.Fatalf("expected not found when updating deleted task, got %v", err)
	}
}
