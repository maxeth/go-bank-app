package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/maxeth/go-bank-app/auth"
)

const (
	authHeaderKey  = "authoriz"
	authTypeBearer = "bearer"
	authPayloadKey = "authorization_payload" // the auth payload will be accessible under this key in gin.Context
)

func authMiddleware(tokenMaker auth.TokenMaker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader((authHeaderKey))
		if len(authHeader) == 0 {
			err := errors.New("authorization header required")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return

		}

		authType := strings.ToLower(fields[0])
		if authType != authTypeBearer {
			err := fmt.Errorf("unsupported authorization type %v", authType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		token := fields[1]
		payload, err := tokenMaker.VerifyToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authPayloadKey, payload) // save the users payload in the context
		ctx.Next()
	}

}
