package middleware

import (
	"Skripsi-Backend/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequireAuth(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Error": "Akses Ditolak Server" + err.Error(),
		})
		c.Abort()
		return
	}
	c.Set("students_uid", claims.ID)
	c.Next()
}
