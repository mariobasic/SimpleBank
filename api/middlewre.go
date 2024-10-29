package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mariobasic/simplebank/token"
	"net/http"
	"strings"
)

const authorizationHeaderKey = "Authorization"
const authorizationTypeBearer = "Bearer"
const authorizationPayloadKey = "authorization_payload"

func authMiddleware(token token.Maker) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(authorizationHeaderKey)
		if len(header) == 0 {
			err := errors.New("authorization header is empty")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(header)
		if len(fields) != 2 || fields[0] != authorizationTypeBearer {
			err := errors.New("invalid authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		payload, err := token.VerifyToken(fields[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		c.Set(authorizationPayloadKey, payload)
		c.Next()
	}

}
