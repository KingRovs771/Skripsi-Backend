package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SavePenyakit(c *gin.Context) {
	var inputPenyakit struct {
		KodePenyakit    string `json:"kode_penyakit" binding:"required"`
		NamaPenyakit    string `json:"nama_penyakit" binding:"required"`
		KodeTurunan     string `json:"kode_turunan"`
		Description     string `json:"description" binding:"required"`
		SaranPenanganan string `json:"saran_penanganan" binding:"required"`
	}

	if err := c.ShouldBindJSON(&inputPenyakit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Pada Server",
			"Error":   err.Error(),
		})
		return
	}

	penyakits := &models.Penyakit{
		KodePenyakit:    inputPenyakit.KodePenyakit,
		NamaPenyakit:    inputPenyakit.NamaPenyakit,
		KodeTurunan:     inputPenyakit.KodeTurunan,
		Description:     inputPenyakit.Description,
		SaranPenanganan: inputPenyakit.SaranPenanganan,
	}

	hasil, err := penyakits.SavePenyakit()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Penyakit save error",
			"Error":   err.Error(),
		})
		return
	}

	// Audit Log (Fitur 3)
	utils.WriteAuditLog(c, "penyakits", hasil.PenyakitUID, "CREATE", nil, hasil)

	c.JSON(http.StatusCreated, gin.H{
		"Status":  "Created",
		"Message": "Data Penyakit Berhasil Disimpan",
		"Data":    hasil,
	})
}

func GetAllPenyakits(c *gin.Context) {
	penyakits, err := models.GetAllPenyakit()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Pada Server",
			"Error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Data Penyakit Berhasil Didapatkan",
		"Data":    penyakits,
	})
}

func GetPenyakitByUID(c *gin.Context) {
	uid := c.Param("uid")
	penyakit, err := models.GetPenyakitByUID(uid)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Server",
			"Error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Data Penyakit Berhasil Didapatkan",
		"Data":    penyakit,
	})
}

func UpdatePenyakit(c *gin.Context) {
	uid := c.Param("uid")

	var inputPenyakitUpdate struct {
		KodePenyakit    string `json:"kode_penyakit" binding:"required"`
		NamaPenyakit    string `json:"nama_penyakit" binding:"required"`
		KodeTurunan     string `json:"kode_turunan"`
		Description     string `json:"description" binding:"required"`
		SaranPenanganan string `json:"saran_penanganan" binding:"required"`
	}

	if err := c.ShouldBindJSON(&inputPenyakitUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Pada Server",
			"Error":   err.Error(),
		})
		return
	}

	var penyakits models.Penyakit

	if err := database.DB.Where("penyakit_uid = ?", uid).First(&penyakits).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Server",
			"Error":   err.Error(),
		})
		return
	}

	// Snapshot data sebelum update (Fitur 3)
	dataSebelum := penyakits

	if inputPenyakitUpdate.KodePenyakit != "" {
		penyakits.KodePenyakit = inputPenyakitUpdate.KodePenyakit
	}

	if inputPenyakitUpdate.NamaPenyakit != "" {
		penyakits.NamaPenyakit = inputPenyakitUpdate.NamaPenyakit
	}

	if inputPenyakitUpdate.KodeTurunan != "" {
		penyakits.KodeTurunan = inputPenyakitUpdate.KodeTurunan
	}

	if inputPenyakitUpdate.Description != "" {
		penyakits.Description = inputPenyakitUpdate.Description
	}

	if inputPenyakitUpdate.SaranPenanganan != "" {
		penyakits.SaranPenanganan = inputPenyakitUpdate.SaranPenanganan
	}

	if err := penyakits.UpdatePenyakit(uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Penyakit save error",
			"Error":   err.Error(),
		})
		return
	}
	// Audit Log (Fitur 3)
	utils.WriteAuditLog(c, "penyakits", uid, "UPDATE", dataSebelum, penyakits)

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Created",
		"Message": "Data Penyakit Berhasil Disimpan",
		"Data":    penyakits,
	})
}

func DeletePenyakit(c *gin.Context) {
	uid := c.Param("uid")

	// Fetch data sebelum delete untuk audit log (Fitur 3)
	var dataSebelum models.Penyakit
	if err := database.DB.Where("penyakit_uid = ?", uid).First(&dataSebelum).Error; err == nil {
		utils.WriteAuditLog(c, "penyakits", uid, "DELETE", dataSebelum, nil)
	}

	err := models.DeletePenyakit(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Server",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{
		"Status":  "Not Found",
		"Message": "Data Penyakit Berhasil DiHapus",
		"Data":    err,
	})

}
