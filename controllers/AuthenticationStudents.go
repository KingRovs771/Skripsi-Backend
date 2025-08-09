package controllers

import (
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterStudents(c *gin.Context) {
	var registerUser struct {
		NISN     string `json:"nisn"`
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := c.BindJSON(&registerUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	RegisterStudents := models.Students{
		Username: registerUser.Username,
		Email:    registerUser.Email,
		Password: registerUser.Password,
	}
	savedStudents, err := RegisterStudents.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"data": savedStudents})
}

func LoginStudents(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loginStudents, err := models.FindUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau password salah"})
		return
	}

	err = loginStudents.ValidatePassword(input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau password salah"})
		return
	}
	jwt, err := utils.GenerateJWTStudents(loginStudents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwt})
}

func GetProfileStudents(c *gin.Context) {
	StudentsUID, exists := c.Get("students_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal mendapatkan user dari context"})
		return
	}

	user, err := models.FindUserByID(StudentsUID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
