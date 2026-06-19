package token

import (
	"time"

	"aidanwoods.dev/go-paseto"
)

const (
	PayloadKey = "payload"
)

type PasetoMaker struct {
	symmetricKey paseto.V4SymmetricKey
}

func NewPasetoMaker(secretKey string) (Maker, error) {
	keyBytes := []byte(secretKey)
	v4SymmetricKey, err := paseto.V4SymmetricKeyFromBytes(keyBytes)
	if err != nil {
		return nil, err
	}
	maker := &PasetoMaker{
		symmetricKey: v4SymmetricKey,
	}
	return maker, nil
}

func (maker *PasetoMaker) CreateToken(id int64, email string, duration time.Duration) (string, *TokenPayload, error) {
	payload, err := NewTokenPayload(id, email, duration)
	if err != nil {
		return "", nil, err
	}
	t := paseto.NewToken()
	t.SetExpiration(payload.ExpiredAt)
	t.SetIssuedAt(payload.IssuedAt)
	t.Set(PayloadKey, payload)
	return t.V4Encrypt(maker.symmetricKey, nil), payload, nil
}

func (maker *PasetoMaker) VerifyToken(tokenString string) (*TokenPayload, error) {
	parser := paseto.NewParser()
	t, err := parser.ParseV4Local(maker.symmetricKey, tokenString, nil)
	if err != nil {
		return nil, err
	}
	payload := TokenPayload{}
	t.Get(PayloadKey, &payload)
	payload.ExpiredAt, _ = t.GetExpiration()
	payload.IssuedAt, _ = t.GetIssuedAt()
	return &payload, nil
}
