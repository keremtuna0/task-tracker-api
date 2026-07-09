package main

import (
	"log"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/keremtuna0/task-tracker-api/internal/config"
	"github.com/keremtuna0/task-tracker-api/internal/database"
	"github.com/keremtuna0/task-tracker-api/internal/httpx"
	"github.com/keremtuna0/task-tracker-api/internal/task"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join("migrations")
	if err := database.Migrate(db, migrationsDir); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	repo := task.NewSQLiteRepository(db)
	service := task.NewService(repo)
	handler := task.NewHandler(service)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return httpx.JSONError(c, fiber.StatusInternalServerError, "internal server error")
		},
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	handler.Register(app)

	log.Printf("starting server on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
