package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func errorHandler(c *fiber.Ctx, err error) error {
	// Default 500 statuscode
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		// Override status code if fiber.Error type
		code = e.Code
	}
	// Set Content-Type: application/json
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	// Return statuscode with error message
	return c.Status(code).JSON(fiber.Map{"error": true, "message": err.Error()})
}

func NotFoundError(message string) *fiber.Error {
	return &fiber.Error{Message: message, Code: http.StatusNotFound}
}

func InternalServerError(message string) *fiber.Error {
	return &fiber.Error{Message: message, Code: http.StatusInternalServerError}
}

func BadRequestError(message string) *fiber.Error {
	return &fiber.Error{Message: message, Code: http.StatusBadRequest}
}

func ValidationError(message string) *fiber.Error {
	return &fiber.Error{Message: message, Code: http.StatusBadRequest}
}
