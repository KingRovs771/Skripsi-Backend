package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Pakar Users
func CreatePakar(c *gin.Context) {
	var input struct {
		RoleUID        int64  `json:"role_uid" binding:"required"`
		NomorSIP       string `json:"nomor_sip" binding:"required,max=20"`
		NamaLengkap    string `json:"nama_lengkap" binding:"required,max=90"`
		JenisSpesialis string `json:"jenis_spesialis" binding:"required"`
		Phone          string `json:"phone" binding:"max=20"`
		Email          string `json:"email" binding:"required,email"`
		Alamat         string `json:"alamat"`
		Password       string `json:"password" binding:"required,min=6"`
		PhotoFile      []byte `json:"photo_file"` // Gambar dalam bentuk byte array
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pakar := models.Pakar{
		RoleUID:        input.RoleUID,
		NomorSIP:       input.NomorSIP,
		NamaLengkap:    input.NamaLengkap,
		JenisSpesialis: input.JenisSpesialis,
		Phone:          input.Phone,
		Email:          input.Email,
		Alamat:         input.Alamat,
		Password:       input.Password,
		PhotoFile:      input.PhotoFile,
	}

	result, err := pakar.SaveUsersPakar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Data pakar berhasil ditambahkan",
		"data":    result,
	})
}
func GetAllPakar(c *gin.Context) {
	pakarList, err := models.GetAllPakar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Hilangkan PhotoFile dari response JSON
	for i := range pakarList {
		pakarList[i].PhotoFile = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data pakar berhasil diambil",
		"data":    pakarList,
	})
}
func GetPakarByUID(c *gin.Context) {
	uid := c.Param("uid")

	pakar, err := models.GetPakarByUID(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pakar tidak ditemukan"})
		return
	}

	// Jangan kirim PhotoFile ke client
	pakar.PhotoFile = nil

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data pakar berhasil diambil",
		"data":    pakar,
	})
}

func UpdatePakar(c *gin.Context) {
	uid := c.Param("uid")

	var input struct {
		RoleUID        *int64  `json:"role_uid,omitempty"`
		NomorSIP       *string `json:"nomor_sip,omitempty" binding:"omitempty,max=20"`
		NamaLengkap    *string `json:"nama_lengkap,omitempty" binding:"omitempty,max=90"`
		JenisSpesialis *string `json:"jenis_spesialis,omitempty"`
		Phone          *string `json:"phone,omitempty" binding:"omitempty,max=20"`
		Email          *string `json:"email,omitempty" binding:"omitempty,email"`
		Alamat         *string `json:"alamat,omitempty"`
		PhotoFile      *[]byte `json:"photo_file,omitempty"` // Optional update photo
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pakar models.Pakar
	if err := database.DB.Where("pakar_uid = ?", uid).First(&pakar).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pakar tidak ditemukan"})
		return
	}

	// Update fields
	if input.RoleUID != nil {
		pakar.RoleUID = *input.RoleUID
	}
	if input.NomorSIP != nil {
		pakar.NomorSIP = *input.NomorSIP
	}
	if input.NamaLengkap != nil {
		pakar.NamaLengkap = *input.NamaLengkap
	}
	if input.JenisSpesialis != nil {
		pakar.JenisSpesialis = *input.JenisSpesialis
	}
	if input.Phone != nil {
		pakar.Phone = *input.Phone
	}
	if input.Email != nil {
		pakar.Email = *input.Email
	}
	if input.Alamat != nil {
		pakar.Alamat = *input.Alamat
	}
	if input.PhotoFile != nil {
		pakar.PhotoFile = *input.PhotoFile
	}

	if err := pakar.UpdatePakar(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Jangan kirim PhotoFile ke client
	pakar.PhotoFile = nil

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data pakar berhasil diperbarui",
		"data":    pakar,
	})
}

func DeletePakar(c *gin.Context) {
	uid := c.Param("uid")

	err := models.DeletePakar(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data pakar berhasil dihapus",
	})
}
