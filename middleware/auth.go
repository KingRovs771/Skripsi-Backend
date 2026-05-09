package middleware

import (
	"Skripsi-Backend/utils"
	_ "Skripsi-Backend/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := utils.ValidateJWT(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Status": "Error", "Message": "Tidak terautentikasi", "Error": err.Error()})
			c.Abort()
			return
		}

		// Simpan data ke context agar bisa digunakan di controller
		c.Set("user_uid", claims.ID)
		c.Set("user_type", claims.UserType)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}

