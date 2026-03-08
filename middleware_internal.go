package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func InternalOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-PORA-KEY")

		if key == "" || key != os.Getenv("PORA_INTERNAL_KEY") {
			c.AbortWithStatusJSON(403, gin.H{
				"error": "forbidden",
			})
			return
		}

		c.Next()
	}
}
