package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SavePertanyaan(c *gin.Context) {
	var inputPertanyan struct {
		KodePertanyaan     string  `json:"kode_pertanyaan" binding:"required"`
		KategoriPertanyaan string  `json:"kategori_pertanyaan" binding:"required"`
		Pertanyaan         string  `json:"pertanyaan" binding:"required"`
		Bobot              float64 `gorm:"type:decimal(10,2)" json:"bobot"`
	}
	if err := c.ShouldBindJSON(&inputPertanyan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Invalid Input Pada Form",
			"Message": "Silakan Ulangi Lagi Input Pertanyaan",
			"Error":   err.Error(),
		})
		return
	}

	quest := &models.Pertanyaan{
		KodePertanyaan:     inputPertanyan.KodePertanyaan,
		KategoriPertanyaan: inputPertanyan.KategoriPertanyaan,
		Pertanyaan:         inputPertanyan.Pertanyaan,
		Bobot:              inputPertanyan.Bobot,
	}

	result, err := quest.SavePertanyaan()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Bad Request Server",
			"Error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"Status":         "Created",
		"Message":        "Pertanyaan Berhasil Disimpan",
		"Data":           result,
		"DEBUG_PAYLOAD":  inputPertanyan,
	})
}

func GetAllPertanyaans(c *gin.Context) {
	quest, err := models.GetAllPertanyaan()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Bad Request Server",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Pertanyaan Berhasil Didapatkan",
		"Data":    quest,
	})
}

func GetPertanyaanByUID(c *gin.Context) {
	uid := c.Param("uid")
	quest, err := models.GetPertanyaanByUID(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Bad Request Server",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Pertanyaan Berhasil Didapatkan",
		"Data":    quest,
	})
}

func UpdatePertanyaan(c *gin.Context) {
	uid := c.Param("uid")

	var inputPertanyan struct {
		KodePertanyaan     string  `json:"kode_pertanyaan" binding:"required"`
		KategoriPertanyaan string  `json:"kategori_pertanyaan" binding:"required"`
		Pertanyaan         string  `json:"pertanyaan" binding:"required"`
		Bobot              float64 `gorm:"type:decimal(10,2)" json:"bobot"`
	}

	if err := c.ShouldBindJSON(&inputPertanyan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Invalid Input Pada Form",
			"Message": "Silakan Ulangi Lagi Input Pertanyaan",
			"Error":   err.Error(),
		})
		return
	}

	var quest models.Pertanyaan

	if err := database.DB.Where("pertanyaan_uid = ?", uid).First(&quest).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Bad Request Server",
			"Error":   err.Error(),
		})
		return
	}

	if inputPertanyan.KodePertanyaan != "" {
		quest.KodePertanyaan = inputPertanyan.KodePertanyaan
	}

	if inputPertanyan.KategoriPertanyaan != "" {
		quest.KategoriPertanyaan = inputPertanyan.KategoriPertanyaan
	}

	if inputPertanyan.Pertanyaan != "" {
		quest.Pertanyaan = inputPertanyan.Pertanyaan
	}

	if inputPertanyan.Bobot != 0 {
		quest.Bobot = inputPertanyan.Bobot
	}

	if err := quest.UpdatePertanyaan(uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Failed to Update Pertanyaan",
			"Error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Pertanyaan Berhasil Diperbarui",
		"Data":    quest,
	})
}

func DeletePertanyaan(c *gin.Context) {
	uid := c.Param("uid")
	err := models.DeletePertanyaan(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Bad Request Server",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Pertanyaan Berhasil Dihapuskan",
		"Data":    err,
	})
}
