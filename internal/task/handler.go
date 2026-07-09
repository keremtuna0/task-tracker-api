package task

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/keremtuna0/task-tracker-api/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(app fiber.Router) {
	app.Post("/tasks", h.Create)
	app.Get("/tasks", h.List)
	app.Get("/tasks/:id", h.GetByID)
	app.Put("/tasks/:id", h.Update)
	app.Delete("/tasks/:id", h.Delete)
}

type createTaskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"due_date"`
}

type updateTaskRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Status      *string  `json:"status"`
	Priority    *string  `json:"priority"`
	DueDate     **string `json:"due_date"`
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req createTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.JSONError(c, fiber.StatusBadRequest, "invalid JSON body")
	}

	task, err := h.service.Create(c.Context(), CreateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	})
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

func (h *Handler) List(c *fiber.Ctx) error {
	tasks, err := h.service.List(c.Context(), ListFilter{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		SortBy:   c.Query("sort"),
		Order:    c.Query("order"),
	})
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(tasks)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return httpx.JSONError(c, fiber.StatusBadRequest, "invalid task id")
	}

	task, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(task)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return httpx.JSONError(c, fiber.StatusBadRequest, "invalid task id")
	}

	var req updateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.JSONError(c, fiber.StatusBadRequest, "invalid JSON body")
	}

	task, err := h.service.Update(c.Context(), id, UpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	})
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(task)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return httpx.JSONError(c, fiber.StatusBadRequest, "invalid task id")
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		return h.handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	if IsValidationError(err) {
		return httpx.JSONError(c, fiber.StatusBadRequest, ValidationMessage(err))
	}
	if errors.Is(err, ErrNotFound) {
		return httpx.JSONError(c, fiber.StatusNotFound, "task not found")
	}

	return httpx.JSONError(c, fiber.StatusInternalServerError, "internal server error")
}

func parseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}
