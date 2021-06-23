package auth

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

// use paseto
type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// creates a new paseto token-manager struct that implements the TokenMaker interface
func NewPasetoMaker(symmetricKey string) (TokenMaker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size. requires key of length %d for the chachapoly algorithm", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}

// creates a new payload including the username, encrypts it and returns the token as a string
func (pm *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	return pm.paseto.Encrypt(pm.symmetricKey, payload, nil)
}

// verifies the token by trying to decrypt it. if successfull, returns the payload of the token, otherwise an error
func (pm *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := pm.paseto.Decrypt(token, pm.symmetricKey, payload, nil)
	if err != nil {
		return nil, err
	}

	// the jwt library would call payload.Valid automatically, with paseto we need to call this method manually
	err = payload.Valid()
	if err != nil {
		return nil, err // token expired
	}

	return payload, nil
}
