package controllers

import (
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func LoginAdmin(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loginAdmin, err := models.FindUserByEmailAdministrator(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau password salah"})
		return
	}

	err = loginAdmin.ValidatePasswordAdministrator(input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau password salah"})
		return
	}
	jwt, err := utils.GenerateJWTAdmin(loginAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwt})
}

func GetProfileAdmin(c *gin.Context) {
	studentsUIDInterface, exists := c.Get("students_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal mendapatkan user dari context"})
		return
	}
	studentsUID, ok := studentsUIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tipe UID user tidak valid"})
		return
	}

	// Panggil fungsi yang benar untuk mencari berdasarkan UID string
	user, err := models.FindUserByID(studentsUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  http.StatusOK,
		"Message": "Data Berhasil Di Dapatkan",
		"Data":    user,
	})
}

func CreateAdministrator(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
}
