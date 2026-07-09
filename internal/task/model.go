package task

import "time"

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"

	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

var (
	validStatuses = map[string]struct{}{
		StatusTodo:       {},
		StatusInProgress: {},
		StatusDone:       {},
	}
	validPriorities = map[string]struct{}{
		PriorityLow:    {},
		PriorityMedium: {},
		PriorityHigh:   {},
	}
)

type Task struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"due_date"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
}

type ListFilter struct {
	Status   string
	Priority string
	SortBy   string
	Order    string
}

type CreateInput struct {
	Title       string
	Description string
	Status      string
	Priority    string
	DueDate     *string
}

type UpdateInput struct {
	Title       *string
	Description *string
	Status      *string
	Priority    *string
	DueDate     **string
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func isValidStatus(status string) bool {
	_, ok := validStatuses[status]
	return ok
}

func isValidPriority(priority string) bool {
	_, ok := validPriorities[priority]
	return ok
}
