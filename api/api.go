package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aashuprogrammer/fee-management-system/db/pgdb"
	"github.com/aashuprogrammer/fee-management-system/token"
	"github.com/aashuprogrammer/fee-management-system/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	app      *fiber.App
	store    pgdb.Store
	valid    *validator.Validate
	config   utils.Config
	token    token.Maker
	cfClient *CashfreeClient
	r2Client *utils.R2Client
}

func NewServer(config utils.Config, store pgdb.Store, tokenMaker token.Maker) (*Server, error) {

	if store == nil {
		return nil, errors.New("store cannot be nil")
	}
	if tokenMaker == nil {
		return nil, errors.New("tokenMaker cannot be nil")
	}
	r2Client, err := utils.NewR2Client(config.R2AccountID, config.R2AccessKeyID, config.R2SecretAccessKey, config.R2BucketName, config.R2PublicURL, "./uploads")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize R2 client: %w", err)
	}

	server := &Server{
		valid:    validator.New(),
		config:   config,
		store:    store,
		token:    tokenMaker,
		cfClient: NewCashfreeClient(config.CashfreeAppID, config.CashfreeSecretKey, config.CashfreeEnvironment),
		r2Client: r2Client,
	}
	server.setupApi()
	return server, nil
}

func (server *Server) Start(port int16) error {
	return server.app.Listen(fmt.Sprintf(":%d", port))
}

type msgResponse struct {
	Msg string `json:"msg"`
}

func (server *Server) setupApi() {
	app := fiber.New(fiber.Config{
		ServerHeader:  "Inflection-Fiber",
		ErrorHandler:  errorHandler,
		BodyLimit:     50 * 1024 * 1024,
		CaseSensitive: true,
	})

	app.Use(logger.New(logger.ConfigDefault))

	app.Use(cors.New())

	app.Use(compress.New())

	// app.Use(csrf.New())

	app.Use(etag.New())

	app.Use(favicon.New())

	// app.Use(limiter.New())

	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "hello"})
	})

	app.Post("/login", server.login)

	// Student Registration endpoints
	app.Post("/register", server.register)
	app.Get("/registrations/:reg_num", server.getRegistration)

	// Admission & Payment endpoints
	app.Get("/courses", server.listCourses)
	app.Post("/admissions", server.createAdmission)
	app.Post("/payments/verify", server.verifyPayment)
	app.Post("/payments/webhook", server.handleWebhook)

	// Admin admissions list endpoint (protected)
	app.Get("/admissions", server.authMiddleware, server.listAdmissions)

	// Serve uploaded files
	app.Static("/uploads", "./uploads")

	// Serve static files from the frontend project folder
	app.Static("/", "../kit-cillege-ui/dist")

	server.app = app
}

func (server *Server) authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
	}

	fields := strings.Fields(authHeader)
	if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization header format"})
	}

	tokenString := fields[1]
	payload, err := server.token.VerifyToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
	}

	c.Locals("user_payload", payload)
	return c.Next()
}
