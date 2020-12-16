package middleware

import (
	"github.com/gin-gonic/gin"
)

func auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()
	}
}
