package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minKeyLen = 32

type JWTMaker struct {
	secretKey string
}

//
func NewJWTMaker(secretKey string) (TokenMaker, error) {
	if len(secretKey) < minKeyLen {
		return nil, fmt.Errorf("invalid key size. key must be at least %d characters", minKeyLen)
	}
	return &JWTMaker{secretKey}, nil
}

func (jm *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload) // token struct
	return token.SignedString([]byte(jm.secretKey))             // signs token struct, turns it into a string
}

// check if input token is valid and return the payload if it is, or an error if it isnt
func (jm *JWTMaker) VerifyToken(token string) (*Payload, error) {

	keyFunc := func(jwtToken *jwt.Token) (interface{}, error) {
		// ensure that the signing method specificed in the token is in fact the one we use (HS256 which is an instance of HMAC)
		_, ok := jwtToken.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			// a token with a false signing method was passed
			return nil, ErrInvalidToken
		}

		// return key for validating
		return []byte(jm.secretKey), nil
	}

	// ParseWithClaims method docs:
	// "Parse, validate, and return a token. keyFunc will receive the parsed token and should return the key for validating. If everything is kosher, err will be nil"
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		errVal, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(errVal.Inner, ErrExpireToken) {
			// the error of type ValidationError thrown when trying to parse the token is of our type ErrExpireToken
			// which means that the parsing failed due to token expiration
			return nil, ErrExpireToken
		}
		// if the token isnt expired, then the token must be invalid
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		// the sent token claims dont match our payload type
		return nil, ErrInvalidToken
	}

	return payload, nil
}
