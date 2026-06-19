package api

import (
	"fmt"
	"time"

	"github.com/aashuprogrammer/fee-management-system/db/pgdb"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type userLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type loginUserResponse struct {
	Token          string    `json:"token"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
	Email          string    `json:"email"`
	ID             int64     `json:"id"`
}

func (server *Server) login(c *fiber.Ctx) error {
	var req userLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	validationErrors := server.validate(req)
	if validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validationErrors)
	}
	users, err := server.store.GetUserByEmail(c.Context(), req.Email)

	if err != nil {
		fmt.Println(pgdb.ErrorCode(err))
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("login not found")
		}
		return InternalServerError(err.Error())
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(req.Password))
	if err != nil {
		return &fiber.Error{Code: fiber.StatusUnauthorized, Message: "password is wrong"}
	}

	token, payload, err := server.token.CreateToken(int64(users.ID), users.Email, server.config.TokenDuration)
	if err != nil {
		return InternalServerError("failed to generate token")
	}
	return c.JSON(loginUserResponse{
		Token:          token,
		TokenExpiresAt: payload.ExpiredAt,
		Email:          payload.Email,
		ID:             payload.ID,
	})
}

