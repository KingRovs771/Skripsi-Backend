package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetSekolah(c *gin.Context) {
	var Sekolah []models.Sekolah
	if err := database.DB.Find(&Sekolah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil data dari database",
			"Error":   err.Error(),
		})
		return
	}

	if len(Sekolah) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Not Found",
			"Message": "Data Sekolah Tidak Ditemukan",
			"Data":    []models.Sekolah{}, // Kembalikan array kosong, bukan string "0"
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Data Sekolah Berhasil Diambil",
		"Data":    Sekolah,
	})
}

func CreateSchool(c *gin.Context) {
	var input struct {
		NPSN          int64  `json:"npsn" binding:"required"`
		NamaSekolah   string `json:"nama_sekolah" binding:"required,max=100"`
		Jenjang       string `json:"jenjang" binding:"required,max=20"`
		AlamatSekolah string `json:"alamat_sekolah"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sekolah := models.Sekolah{
		NPSN:          input.NPSN,
		NamaSekolah:   input.NamaSekolah,
		Jenjang:       input.Jenjang,
		AlamatSekolah: input.AlamatSekolah,
	}

	result, err := sekolah.SaveSekolah()
	if err != nil {
		// Cek apakah errornya karena NPSN duplikat
		if err.Error() == "NPSN tersebut sudah terdaftar di sistem" {
			c.JSON(http.StatusConflict, gin.H{
				"status":  http.StatusConflict,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Data sekolah berhasil ditambahkan",
		"data":    result,
	})
}
func GetSekolahByUID(c *gin.Context) {
	uid := c.Param("uid")

	sekolah, err := models.GetSekolahByUID(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sekolah tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data sekolah berhasil diambil",
		"data":    sekolah,
	})
}
func UpdateSekolah(c *gin.Context) {
	uid := c.Param("uid")

	var input struct {
		NPSN          *int64  `json:"npsn" binding:"required"`
		NamaSekolah   *string `json:"nama_sekolah" binding:"required,max=100"`
		Jenjang       *string `json:"jenjang" binding:"required,max=20"`
		AlamatSekolah *string `json:"alamat_sekolah"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if input.NPSN != nil {
		updates["npsn"] = *input.NPSN
	}
	if input.NamaSekolah != nil {
		updates["nama_sekolah"] = *input.NamaSekolah
	}
	if input.Jenjang != nil {
		updates["jenjang"] = *input.Jenjang
	}
	if input.AlamatSekolah != nil {
		updates["alamat_sekolah"] = *input.AlamatSekolah
	}

	var sekolah models.Sekolah
	if err := database.DB.Where("sekolah_uid = ?", uid).First(&sekolah).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sekolah tidak ditemukan"})
		return
	}

	if err := sekolah.UpdateSekolah(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data sekolah berhasil diperbarui",
		"Data":    sekolah,
	})
}
func DeleteSekolah(c *gin.Context) {
	uid := c.Param("uid")

	err := models.DeleteSekolah(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data sekolah berhasil dihapus",
	})
}
