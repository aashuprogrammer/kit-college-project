package api

import (
	"github.com/gofiber/fiber/v2"
)

func (server *Server) listCourses(c *fiber.Ctx) error {
	courses, err := server.store.ListCourses(c.Context())
	if err != nil {
		return InternalServerError("failed to retrieve courses: " + err.Error())
	}
	return c.JSON(courses)
}
