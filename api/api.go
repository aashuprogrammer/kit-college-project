package api

import (
	"errors"
	"fmt"

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
}

func NewServer(config utils.Config, store pgdb.Store, tokenMaker token.Maker) (*Server, error) {

	if store == nil {
		return nil, errors.New("store cannot be nil")
	}
	if tokenMaker == nil {
		return nil, errors.New("tokenMaker cannot be nil")
	}
	server := &Server{
		valid:    validator.New(),
		config:   config,
		store:    store,
		token:    tokenMaker,
		cfClient: NewCashfreeClient(config.CashfreeAppID, config.CashfreeSecretKey, config.CashfreeEnvironment),
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
		BodyLimit:     2 * 1024 * 1024,
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

	// Admission & Payment endpoints
	app.Get("/courses", server.listCourses)
	app.Post("/admissions", server.createAdmission)
	app.Post("/payments/verify", server.verifyPayment)
	app.Post("/payments/webhook", server.handleWebhook)

	server.app = app
}
