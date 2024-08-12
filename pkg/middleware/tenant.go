package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		c.Request = c.Request.WithContext(context.WithValue(ctx, "tenant", c.Request.Host))
		c.Next()
	}
}
