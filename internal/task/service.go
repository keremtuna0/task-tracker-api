package task

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, validationError("title is required")
	}

	status := input.Status
	if status == "" {
		status = StatusTodo
	}
	if !isValidStatus(status) {
		return nil, validationError("invalid status")
	}

	priority := input.Priority
	if priority == "" {
		priority = PriorityMedium
	}
	if !isValidPriority(priority) {
		return nil, validationError("invalid priority")
	}

	now := nowUTC()
	task := &Task{
		Title:       title,
		Description: input.Description,
		Status:      status,
		Priority:    priority,
		DueDate:     input.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Task, error) {
	if filter.Status != "" && !isValidStatus(filter.Status) {
		return nil, validationError("invalid status filter")
	}
	if filter.Priority != "" && !isValidPriority(filter.Priority) {
		return nil, validationError("invalid priority filter")
	}
	if filter.SortBy != "" {
		if _, ok := sortColumns[filter.SortBy]; !ok {
			return nil, validationError("invalid sort field")
		}
	}
	if filter.Order != "" {
		if _, ok := sortOrders[filter.Order]; !ok {
			return nil, validationError("invalid sort order")
		}
	}

	tasks, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		return []Task{}, nil
	}

	return tasks, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, validationError("title cannot be empty")
		}
		task.Title = title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Status != nil {
		if !isValidStatus(*input.Status) {
			return nil, validationError("invalid status")
		}
		task.Status = *input.Status
	}
	if input.Priority != nil {
		if !isValidPriority(*input.Priority) {
			return nil, validationError("invalid priority")
		}
		task.Priority = *input.Priority
	}
	if input.DueDate != nil {
		task.DueDate = *input.DueDate
	}

	task.UpdatedAt = nowUTC()

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	deletedAt := nowUTC()
	return s.repo.SoftDelete(ctx, id, deletedAt)
}

func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}

func ValidationMessage(err error) string {
	var validationErr ValidationError
	if errors.As(err, &validationErr) {
		return validationErr.Message
	}
	return "invalid input"
}
