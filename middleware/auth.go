package middleware

import (
	"Skripsi-Backend/models"
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

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	students, err := models.FindUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau password salah"})
		return
	}

	// Verifikasi password yang diinput dengan hash di database.
	err = students.ValidatePassword(input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau password salah"})
		return
	}

	// Jika berhasil, buat token JWT.
	jwt, err := utils.GenerateJWTStudents(students)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwt})
}
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal mendapatkan user dari context"})
		return
	}

	// Cari user berdasarkan ID yang didapat dari token.
	user, err := models.FindUserByID(userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
