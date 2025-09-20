package middleware

import "github.com/gin-gonic/gin"

// Logger returns Gin's default structured logger.
func Logger() gin.HandlerFunc { return gin.Logger() }
