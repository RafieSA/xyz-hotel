package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"

	"xyz-hotel/backend/internal/handler"
	"xyz-hotel/backend/internal/middleware"
	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"
	"xyz-hotel/backend/internal/service"
)
func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-super-secret"
		slog.Warn("JWT_SECRET empty, using default (not for production)")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/xyz_hotel?sslmode=disable"
	}

	var db *sqlx.DB
	var err error
	db, err = sqlx.Connect("pgx", databaseURL)
	if err != nil {
		slog.Error("db connect failed", "err", err, "url", databaseURL)
	} else {
		slog.Info("db connected")
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		// Expired ticker: every 5m UPDATE bookings SET status='expired' WHERE status='pending_payment' AND created_at < now() - interval '12 hours'
		bookingRepo := repo.NewBookingRepo(db)
		go func() {
			// initial run after 10s
			time.Sleep(10 * time.Second)
			if n, err := bookingRepo.ExpirePending(); err != nil {
				slog.Error("expire ticker failed", "err", err)
			} else if n > 0 {
				slog.Info("expired bookings", "count", n)
			}
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				n, err := bookingRepo.ExpirePending()
				if err != nil {
					slog.Error("expire ticker failed", "err", err)
				} else if n > 0 {
					slog.Info("expired bookings", "count", n)
				}
			}
		}()
	}

	// Ensure upload dir exists (private storage)
	if err := os.MkdirAll(filepath.Join("storage", "uploads", "bookings"), 0755); err != nil {
		slog.Error("failed to create upload dir", "err", err)
	}

	app := fiber.New(fiber.Config{
		AppName:   "xyz-hotel",
		BodyLimit: 6 * 1024 * 1024, // 6MB (proof limit 5MB + overhead)
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowCredentials: false,
	}))

	// Health checks (public)
	app.Get("/health", handler.Health)
	app.Get("/api/health", handler.Health)

	var bookingHandler *handler.BookingHandler
	var authHandler *handler.AuthHandler
	var voucherHandler *handler.VoucherHandler
	var reviewHandler *handler.ReviewHandler
	var reportHandler *handler.ReportHandler
	if db != nil {
		userRepo := repo.NewUserRepo(db)
		roomRepo := repo.NewRoomRepo(db)
		bookingRepo := repo.NewBookingRepo(db)
		voucherRepo := repo.NewVoucherRepo(db)
		_ = roomRepo
		authSvc := service.NewAuthService(userRepo, jwtSecret)
		authHandler = handler.NewAuthHandler(authSvc)
		availSvc := service.NewAvailabilityService(db)
		// ensure AvailabilityService has repos set (if constructor doesn't set)
		if availSvc.BookingRepo == nil {
			availSvc.BookingRepo = bookingRepo
		}
		if availSvc.RoomRepo == nil {
			availSvc.RoomRepo = roomRepo
		}
		if availSvc.VoucherRepo == nil {
			availSvc.VoucherRepo = voucherRepo
		}
		voucherSvc := service.NewVoucherService(voucherRepo, roomRepo)
		voucherHandler = handler.NewVoucherHandler(voucherSvc)
		opsSvc := service.NewBookingOpsService(db, bookingRepo, roomRepo)
		bookingHandler = handler.NewBookingHandlerWithOps(availSvc, bookingRepo, opsSvc)
		reviewRepo := repo.NewReviewRepo(db)
		reviewSvc := service.NewReviewService(reviewRepo, bookingRepo, roomRepo)
		reviewHandler = handler.NewReviewHandler(reviewSvc)
		reportSvc := service.NewReportService(db)
		reportHandler = handler.NewReportHandler(reportSvc)
	}

	// Public auth routes
	if authHandler != nil {
		auth := app.Group("/api/auth")
		auth.Post("/register", authHandler.Register)
		auth.Post("/login", authHandler.Login)
		auth.Post("/refresh", authHandler.Refresh)
		auth.Get("/me", middleware.Auth(jwtSecret), authHandler.Me)
	}

	// Public availability + public voucher validation + reviews
	if bookingHandler != nil {
		app.Get("/api/availability", bookingHandler.GetAvailability)
		app.Get("/api/room-types", func(c *fiber.Ctx) error {
			rr := repo.NewRoomRepo(db)
			list, err := rr.ListRoomTypesWithRating()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load room types. Please try again"})
			}
			return c.JSON(fiber.Map{"data": list})
		})
		if reviewHandler != nil {
			app.Get("/api/reviews", reviewHandler.ListReviews)
			app.Post("/api/reviews", middleware.Auth(jwtSecret), reviewHandler.CreateReview)
		}
		if voucherHandler != nil {
			app.Get("/api/vouchers/validate", voucherHandler.ValidateVoucher)
		}
	} else {
		app.Get("/api/availability", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"message": "Service is temporarily unavailable. Please try again later"})
		})
	}

	// Bookings (auth required)
	if bookingHandler != nil {
		bookings := app.Group("/api/bookings", middleware.Auth(jwtSecret))
		bookings.Get("/", bookingHandler.ListBookings)
		bookings.Post("/", bookingHandler.CreateBooking)
		bookings.Post("/:id/proof", bookingHandler.UploadProof)
		app.Get("/api/bookings", middleware.Auth(jwtSecret), bookingHandler.ListBookings)
		app.Post("/api/bookings", middleware.Auth(jwtSecret), bookingHandler.CreateBooking)
		app.Post("/api/bookings/:id/proof", middleware.Auth(jwtSecret), bookingHandler.UploadProof)
	} else {
		booking := app.Group("/api/bookings")
		booking.Get("/", handler.ListBookingsStub)
		booking.Post("/", handler.CreateBookingStub)
	}

	// Admin (owner/manager only) if db available
	if db != nil && bookingHandler != nil {
		admin := app.Group("/api/admin", middleware.Auth(jwtSecret), middleware.RequireRole(model.RoleOwner, model.RoleManager))
		admin.Get("/ping", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "Admin access confirmed", "role": c.Locals("role")})
		})
		admin.Get("/bookings", bookingHandler.ListBookings)
		admin.Patch("/bookings/:id/verify", bookingHandler.VerifyBooking)
		if voucherHandler != nil {
			admin.Get("/vouchers", voucherHandler.ListVouchers)
			admin.Post("/vouchers", voucherHandler.CreateVoucher)
		}
		if reportHandler != nil {
			admin.Get("/reports/summary", reportHandler.GetSummary)
			admin.Get("/reports/revenue", reportHandler.GetRevenue)
			admin.Get("/reports/occupancy", reportHandler.GetOccupancy)
		}
		admin.Get("/rooms", func(c *fiber.Ctx) error {
			rr := repo.NewRoomRepo(db)
			types, err := rr.ListRoomTypes()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load rooms. Please try again"})
			}
			return c.JSON(fiber.Map{"data": types})
		})
		// Ops: check-in/out + room unit status allowed for owner/manager/receptionist
		ops := app.Group("/api/admin", middleware.Auth(jwtSecret), middleware.RequireRole(model.RoleOwner, model.RoleManager, model.RoleReceptionist))
		ops.Patch("/bookings/:id/checkin", bookingHandler.CheckIn)
		ops.Patch("/bookings/:id/checkout", bookingHandler.CheckOut)
		ops.Patch("/room-units/:id/status", bookingHandler.UpdateRoomUnitStatus)
		ops.Get("/room-units", func(c *fiber.Ctx) error {
			rr := repo.NewRoomRepo(db)
			list, err := rr.ListAllUnits(c.Context())
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load room units. Please try again"})
			}
			return c.JSON(fiber.Map{"data": list})
		})
	}
	slog.Info("server starting", "port", port, "service", "xyz-hotel")
	if err := app.Listen(":" + port); err != nil {
		slog.Error("listen failed", "err", err)
		os.Exit(1)
	}
}
