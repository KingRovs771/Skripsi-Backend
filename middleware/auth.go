package middleware

import (
	"Skripsi-Backend/utils"
	_ "Skripsi-Backend/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := utils.ValidateJWT(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}
