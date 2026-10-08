package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"xyz-hotel/backend/internal/handler"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := fiber.New(fiber.Config{
		AppName: "xyz-hotel",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowCredentials: false,
	}))

	// Health checks
	app.Get("/health", handler.Health)
	app.Get("/api/health", handler.Health)

	// Stub booking routes (Fase 1)
	booking := app.Group("/api/bookings")
	booking.Get("/", handler.ListBookingsStub)
	booking.Post("/", handler.CreateBookingStub)

	slog.Info("server starting", "port", port, "service", "xyz-hotel")
	if err := app.Listen(":" + port); err != nil {
		slog.Error("listen failed", "err", err)
		os.Exit(1)
	}
}
