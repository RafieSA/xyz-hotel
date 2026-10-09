package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/websocket/v2"
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
			expireAndNotify := func() {
				// Fetch candidates that will expire (for email + audit) before UPDATE
				type candidate struct {
					ID     int64  `db:"id"`
					UserID int64  `db:"user_id"`
					Email  string `db:"email"`
				}
				var cands []candidate
				_ = db.Select(&cands, `SELECT b.id, b.user_id, u.email FROM bookings b JOIN users u ON u.id=b.user_id WHERE b.status='pending_payment' AND b.created_at < now() - interval '12 hours' LIMIT 100`)
				if n, err := bookingRepo.ExpirePending(); err != nil {
					slog.Error("expire ticker failed", "err", err)
				} else {
					if n > 0 {
						slog.Info("expired bookings", "count", n)
					}
					// Audit + email per expired booking (non-blocking email)
					for _, c := range cands {
						_, _ = db.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`, c.UserID, "booking.expired", "bookings", c.ID, `{"reason":"auto-expired after 12h"}`)
						to := c.Email
						if to == "" {
							to = "unknown@xyz-hotel.local"
						}
						subject := service.BookingExpiredSubject(c.ID)
						body := service.BookingExpiredBody(c.ID)
						service.SendAsync(to, subject, body)
					}
				}
			}
			expireAndNotify()
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				expireAndNotify()
			}
		}()
	}

	// Ensure upload dir exists (private storage)
	if err := os.MkdirAll(filepath.Join("storage", "uploads", "bookings"), 0755); err != nil {
		slog.Error("failed to create upload dir", "err", err)
	}
	// Ensure email log dir exists
	if err := os.MkdirAll(filepath.Join("logs"), 0755); err != nil {
		slog.Error("failed to create logs dir", "err", err)
	}
	if err := os.MkdirAll(filepath.Join("storage", "invoices"), 0755); err != nil {
		slog.Error("failed to create invoices dir", "err", err)
	}
	if err := os.MkdirAll(filepath.Join("backend", "logs"), 0755); err != nil {
		slog.Error("failed to create backend/logs dir", "err", err)
	}
	app := fiber.New(fiber.Config{
		AppName:   "xyz-hotel",
		BodyLimit: 6 * 1024 * 1024, // 6MB (proof limit 5MB + overhead)
	})
	app.Static("/storage", "./storage")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowCredentials: false,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 60 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		Next: func(c *fiber.Ctx) bool {
			p := c.Path()
			return p == "/health" || p == "/api/health"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"message": "Too many requests, try in 60s"})
		},
	}))

	// Health checks (public)
	app.Get("/health", handler.Health)
	app.Get("/api/health", handler.Health)

	var bookingHandler *handler.BookingHandler
	var authHandler *handler.AuthHandler
	var voucherHandler *handler.VoucherHandler
	var reviewHandler *handler.ReviewHandler
	var reportHandler *handler.ReportHandler
	var auditHandler *handler.AuditHandler
	var wishlistHandler *handler.WishlistHandler
	var roomHandler *handler.RoomHandler
	var invoiceHandler *handler.InvoiceHandler
	var roomImageHandler *handler.RoomImageHandler
	var addonHandler *handler.AddonHandler
	var loyaltyHandler *handler.LoyaltyHandler
	var chatHandler *handler.ChatHandler
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
		auditRepo := repo.NewAuditRepo(db)
		auditHandler = handler.NewAuditHandler(auditRepo)
		wishlistRepo := repo.NewWishlistRepo(db)
		wishlistSvc := service.NewWishlistService(wishlistRepo)
		wishlistHandler = handler.NewWishlistHandler(wishlistSvc)
		roomHandler = handler.NewRoomHandler(roomRepo)
		invoiceSvc := service.NewInvoiceService(db)
		invoiceHandler = handler.NewInvoiceHandler(invoiceSvc, bookingRepo)
		// Gallery
		roomImageRepo := repo.NewRoomImageRepo(db)
		roomImageHandler = handler.NewRoomImageHandler(roomImageRepo, roomRepo)
		// Addons
		addonRepo := repo.NewAddonRepo(db)
		addonSvc := service.NewAddonService(db, addonRepo, bookingRepo)
		addonHandler = handler.NewAddonHandler(addonSvc)
		// Loyalty
		loyaltySvc := service.NewLoyaltyService(db)
		loyaltyHandler = handler.NewLoyaltyHandler(loyaltySvc)
		// Chat
		messageRepo := repo.NewMessageRepo(db)
		chatHandler = handler.NewChatHandler(messageRepo, bookingRepo)
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
		if roomHandler != nil {
			app.Get("/api/room-types", roomHandler.ListRoomTypesPublic)
		} else {
			app.Get("/api/room-types", func(c *fiber.Ctx) error {
				rr := repo.NewRoomRepo(db)
				list, err := rr.ListRoomTypesWithRating()
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load room types. Please try again"})
				}
				return c.JSON(fiber.Map{"data": list})
			})
		}
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
	// Gallery public + addons public
	if roomImageHandler != nil {
		app.Get("/api/room-types/:id/images", roomImageHandler.ListImages)
	}
	if addonHandler != nil {
		app.Get("/api/addons", addonHandler.ListAddons)
	}
	// Bookings (auth required)
	if bookingHandler != nil {
		bookings := app.Group("/api/bookings", middleware.Auth(jwtSecret))
		bookings.Get("/", bookingHandler.ListBookings)
		bookings.Post("/", bookingHandler.CreateBooking)
		bookings.Post("/:id/proof", bookingHandler.UploadProof)
		bookings.Patch("/:id/cancel", bookingHandler.CancelBooking)
		if addonHandler != nil {
			bookings.Post("/:id/addons", addonHandler.AddToBooking)
			bookings.Get("/:id/addons", addonHandler.ListBookingAddons)
		}
		if invoiceHandler != nil {
			bookings.Get("/:id/invoice", invoiceHandler.GetInvoice)
		}
		app.Get("/api/bookings", middleware.Auth(jwtSecret), bookingHandler.ListBookings)
		app.Post("/api/bookings", middleware.Auth(jwtSecret), bookingHandler.CreateBooking)
		app.Post("/api/bookings/:id/proof", middleware.Auth(jwtSecret), bookingHandler.UploadProof)
		app.Patch("/api/bookings/:id/cancel", middleware.Auth(jwtSecret), bookingHandler.CancelBooking)
		if addonHandler != nil {
			app.Post("/api/bookings/:id/addons", middleware.Auth(jwtSecret), addonHandler.AddToBooking)
			app.Get("/api/bookings/:id/addons", middleware.Auth(jwtSecret), addonHandler.ListBookingAddons)
		}
		if invoiceHandler != nil {
			app.Get("/api/bookings/:id/invoice", middleware.Auth(jwtSecret), invoiceHandler.GetInvoice)
		}
	} else {
		booking := app.Group("/api/bookings")
		booking.Get("/", handler.ListBookingsStub)
		booking.Post("/", handler.CreateBookingStub)
	}
	// Loyalty (auth required)
	if loyaltyHandler != nil {
		app.Get("/api/loyalty/points", middleware.Auth(jwtSecret), loyaltyHandler.GetPoints)
		app.Post("/api/loyalty/redeem", middleware.Auth(jwtSecret), loyaltyHandler.Redeem)
	}
	// Wishlist (auth required) - heart toggle on room cards
	if wishlistHandler != nil {
		wl := app.Group("/api/wishlist", middleware.Auth(jwtSecret))
		wl.Get("/", wishlistHandler.List)
		wl.Post("/toggle", wishlistHandler.Toggle)
		wl.Delete("/:room_type_id", wishlistHandler.Delete)
		// also support without trailing slash
		app.Get("/api/wishlist", middleware.Auth(jwtSecret), wishlistHandler.List)
		app.Post("/api/wishlist/toggle", middleware.Auth(jwtSecret), wishlistHandler.Toggle)
		app.Delete("/api/wishlist/:room_type_id", middleware.Auth(jwtSecret), wishlistHandler.Delete)
	}
	// Chat (auth required)
	if chatHandler != nil {
		chat := app.Group("/api/chat", middleware.Auth(jwtSecret))
		chat.Post("/", chatHandler.PostChat)
		chat.Get("/", chatHandler.GetChat)
		chat.Get("/conversations", chatHandler.GetConversations)
		// also without trailing slash
		app.Post("/api/chat", middleware.Auth(jwtSecret), chatHandler.PostChat)
		app.Get("/api/chat", middleware.Auth(jwtSecret), chatHandler.GetChat)
		app.Get("/api/chat/conversations", middleware.Auth(jwtSecret), chatHandler.GetConversations)
		// Admin inbox
		app.Get("/api/admin/chat/messages", middleware.Auth(jwtSecret), middleware.RequireRole(model.RoleOwner, model.RoleManager, model.RoleReceptionist), chatHandler.GetAdminMessages)
	}

	// WebSocket admin realtime (Auth+RBAC) — must be before admin group to allow upgrade
	if db != nil && bookingHandler != nil {
		app.Get("/ws/admin", middleware.Auth(jwtSecret), middleware.RequireRole(model.RoleOwner, model.RoleManager, model.RoleReceptionist), handler.WsAdmin, websocket.New(handler.WsAdminHandler))
	}
	// WebSocket chat realtime (Auth, any role) shares same hub for broadcast
	if db != nil && chatHandler != nil {
		app.Get("/ws/chat", middleware.Auth(jwtSecret), handler.WsChat, websocket.New(handler.WsChatHandler))
	}

	// Admin (owner/manager only) if db available
	if db != nil && bookingHandler != nil {
		admin := app.Group("/api/admin", middleware.Auth(jwtSecret), middleware.RequireRole(model.RoleOwner, model.RoleManager))
		admin.Get("/ping", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "Admin access confirmed", "role": c.Locals("role")})
		})
		admin.Get("/bookings", bookingHandler.ListBookings)
		admin.Patch("/bookings/:id/verify", bookingHandler.VerifyBooking)
		if auditHandler != nil {
			admin.Get("/audit-logs", auditHandler.ListAuditLogs)
		}
		if voucherHandler != nil {
			admin.Get("/vouchers", voucherHandler.ListVouchers)
			admin.Post("/vouchers", voucherHandler.CreateVoucher)
		}
		if reportHandler != nil {
			admin.Get("/reports/summary", reportHandler.GetSummary)
			admin.Get("/reports/revenue", reportHandler.GetRevenue)
			admin.Get("/reports/occupancy", reportHandler.GetOccupancy)
			admin.Get("/reports/export.csv", reportHandler.ExportCSV)
		}
		admin.Get("/rooms", func(c *fiber.Ctx) error {
			rr := repo.NewRoomRepo(db)
			types, err := rr.ListRoomTypes()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load rooms. Please try again"})
			}
			return c.JSON(fiber.Map{"data": types})
		})
		if roomHandler != nil {
			admin.Get("/room-types", roomHandler.ListRoomTypesAdmin)
			admin.Post("/room-types", roomHandler.CreateRoomType)
			admin.Put("/room-types/:id", roomHandler.UpdateRoomType)
			admin.Delete("/room-types/:id", roomHandler.DeleteRoomType)
			admin.Get("/room-units", roomHandler.ListRoomUnits)
			admin.Post("/room-units", roomHandler.CreateRoomUnit)
			admin.Delete("/room-units/:id", roomHandler.DeleteRoomUnit)
		}
		if roomImageHandler != nil {
			admin.Post("/room-types/:id/images", roomImageHandler.UploadImage)
			admin.Delete("/room-types/:id/images/:imageId", roomImageHandler.DeleteImage)
			admin.Delete("/room-images/:id", roomImageHandler.DeleteImage)
		}
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
		if reportHandler != nil {
			ops.Get("/calendar", reportHandler.GetCalendar)
		}
	}
	slog.Info("server starting", "port", port, "service", "xyz-hotel")
	if err := app.Listen(":" + port); err != nil {
		slog.Error("listen failed", "err", err)
		os.Exit(1)
	}
}
