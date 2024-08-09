package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		host := c.Request.Host

		c.Request = c.Request.WithContext(context.WithValue(ctx, "tenant", strings.Split(host, ".")[0]))
		c.Next()
	}
}
