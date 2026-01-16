package controllers

import (
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterStudents(c *gin.Context) {
	var registerUser struct {
		NISN              string `json:"nisn"`
		NamaLengkap       string `json:"nama_lengkap"`
		NamaInisial       string `json:"nama_inisial"`
		JenjangPendidikan string `json:"jenjang_pendidikan"`
		Kelas             int64  `json:"kelas"`
		Username          string `json:"username"`
		Password          string `json:"password"`
		Email             string `json:"email"`
	}

	if err := c.BindJSON(&registerUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	RegisterStudents := models.Students{
		NISN:              registerUser.NISN,
		NamaLengkap:       registerUser.NamaLengkap,
		JenjangPendidikan: registerUser.JenjangPendidikan,
		Kelas:             registerUser.Kelas,
		Password:          registerUser.Password,
		Email:             registerUser.Email,
	}
	savedStudents, err := RegisterStudents.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  http.StatusOK,
		"Message": "Data Register Telah Berhasil di Simpan, Silakan Lanjutkan Proses Login",
		"data":    savedStudents,
	})
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
