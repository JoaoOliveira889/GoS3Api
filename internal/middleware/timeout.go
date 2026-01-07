package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestTimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{}, 1)

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			return
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
					"error": "request timed out",
				})
			}
		}
	}
}
