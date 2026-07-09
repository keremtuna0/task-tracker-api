package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, task *Task) error
	List(ctx context.Context, filter ListFilter) ([]Task, error)
	GetByID(ctx context.Context, id int64) (*Task, error)
	Update(ctx context.Context, task *Task) error
	SoftDelete(ctx context.Context, id int64, deletedAt string) error
}

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

var sortColumns = map[string]string{
	"created_at": "created_at",
	"due_date":   "due_date",
}

var sortOrders = map[string]string{
	"asc":  "ASC",
	"desc": "DESC",
}

func (r *SQLiteRepository) Create(ctx context.Context, task *Task) error {
	query := `
		INSERT INTO tasks (title, description, status, priority, due_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}

	task.ID = id
	return nil
}

func (r *SQLiteRepository) List(ctx context.Context, filter ListFilter) ([]Task, error) {
	sortBy := sortColumns[filter.SortBy]
	if sortBy == "" {
		sortBy = "created_at"
	}

	order := sortOrders[filter.Order]
	if order == "" {
		order = "DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, status, priority, due_date, created_at, updated_at
		FROM tasks
		WHERE deleted_at IS NULL
	`)

	args := make([]any, 0, 2)
	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.Priority != "" {
		query += " AND priority = ?"
		args = append(args, filter.Priority)
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, order)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.Priority,
			&t.DueDate,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id int64) (*Task, error) {
	query := `
		SELECT id, title, description, status, priority, due_date, created_at, updated_at
		FROM tasks
		WHERE id = ? AND deleted_at IS NULL
	`

	var t Task
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.Priority,
		&t.DueDate,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	return &t, nil
}

func (r *SQLiteRepository) Update(ctx context.Context, task *Task) error {
	query := `
		UPDATE tasks
		SET title = ?, description = ?, status = ?, priority = ?, due_date = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
		task.UpdatedAt,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *SQLiteRepository) SoftDelete(ctx context.Context, id int64, deletedAt string) error {
	query := `
		UPDATE tasks
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, deletedAt, deletedAt, id)
	if err != nil {
		return fmt.Errorf("soft delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
