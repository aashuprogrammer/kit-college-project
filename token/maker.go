package token

import (
	"time"
)

type Maker interface {
	CreateToken(id int64, email string, duration time.Duration) (string, *TokenPayload, error)
	VerifyToken(token string) (*TokenPayload, error)
}

type TokenPayload struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewTokenPayload(id int64, email string, duration time.Duration) (*TokenPayload, error) {
	return &TokenPayload{
		ID:        id,
		Email:     email,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}, nil
}
