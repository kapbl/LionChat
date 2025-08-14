package middlewares

import (
	"cchat/pkg/token"

	"github.com/gin-gonic/gin"
)

const (
	XRequestIDKey = "X-Request-ID"
)

// RequestID is a middleware that injects a request ID into the context of each request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := token.NewRequestID()
		// Set the request ID in the context
		c.Set("request_id", requestID)

		// Set the request ID in the response header
		c.Header(XRequestIDKey, requestID)

		// Continue to the next middleware or handler
		c.Next()
	}
}
