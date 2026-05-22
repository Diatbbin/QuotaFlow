package token

import (
	"errors"
	"time"
	"github.com/o1egl/paseto"
	"github.com/aead/chacha20poly1305"
)

type PasetoMaker struct {
	paseto *paseto.V2
	symmetricKey []byte
}

func NewPasetoMaker(symetricKey string) (*PasetoMaker, error) {
	if len(symetricKey) != chacha20poly1305.KeySize {
		return nil, errors.New("invalid key size: must be exactly 32 bytes")
	}

	return &PasetoMaker{
		paseto: paseto.NewV2(),
		symmetricKey: []byte(symetricKey),
	}, nil
}

func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, errors.New("token is invalid")
	}

	err = payload.Validate(); if err != nil {
		return nil, err
	}

	return payload, nil
}


func (payload *Payload) Validate() error {
	if time.Now().After(payload.ExpiredAt) {
		return errors.New("token has expired")
	}

	return nil
}