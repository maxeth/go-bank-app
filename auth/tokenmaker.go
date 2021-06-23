package auth

import "time"

type TokenMaker interface {
	// create a token
	CreateToken(username string, duration time.Duration) (string, error)
	// check if input token is valid and return its payload if so
	VerifyToken(token string) (*Payload, error)
}
