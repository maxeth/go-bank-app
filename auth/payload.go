package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExpireToken  = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issuedAt"`
	ExpiredAt time.Time `json:"expiredAt"`
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

// checks whether payload is valid aka., whether token isn't expired
func (pl *Payload) Valid() error {
	if time.Now().After(pl.ExpiredAt) {
		return ErrExpireToken
	}

	return nil
}
