package middleware

import "github.com/gin-gonic/gin"

// Recover adds panic recovery with a 500 if something goes wrong.
func Recover() gin.HandlerFunc { return gin.Recovery() }
