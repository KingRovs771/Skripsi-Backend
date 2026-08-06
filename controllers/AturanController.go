package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllAturan(c *gin.Context) {
	Aturan, err := models.GetAllAturan()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Bad Request Server",
			"Error":   err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Aturan Successfully Get",
		"Data":    Aturan,
	})
}
func SaveAturan(c *gin.Context) {
	var InputAturan struct {
		AturanUID                string `json:"aturan_uid"`
		KodePenyakit             string `json:"kode_penyakit" binding:"required"`
		KodePertanyaan           string `json:"kode_pertanyaan" binding:"required"`
		MinValue                 int64  `json:"min_value"`
		IsMandatory              int64  `json:"is_mandatory"`
		TipeAturan               string `json:"tipe_aturan"`
		BerlakuUntukSemuaTingkat bool   `json:"berlaku_untuk_semua_tingkat"`
	}

	if err := c.ShouldBindJSON(&InputAturan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Silakan Ulangi Lagi Input Aturan",
			"Error":   err,
		})
		return
	}

	tipe := InputAturan.TipeAturan
	if tipe == "" {
		tipe = "GEJALA_INTI"
	}

	aturan := models.Aturan{
		AturanUID:                InputAturan.AturanUID,
		KodePenyakit:             InputAturan.KodePenyakit,
		KodePertanyaan:           InputAturan.KodePertanyaan,
		MinValue:                 InputAturan.MinValue,
		IsMandatory:              InputAturan.IsMandatory,
		TipeAturan:               tipe,
		BerlakuUntukSemuaTingkat: InputAturan.BerlakuUntukSemuaTingkat,
	}

	result, err := aturan.SaveAturan()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Silakan Ulangi Lagi Input Aturan",
			"Error":   err,
		})
		return
	}
	// Audit Log (Fitur 3)
	utils.WriteAuditLog(c, "aturans", result.AturanUID, "CREATE", nil, result)

	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Aturan Successfully Saved",
		"Data":    result,
	})
}
func GetAturanByUID(c *gin.Context) {
	uid := c.Param("uid")
	aturan, err := models.GetAturanByUID(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Data Aturan Not Found",
			"Error":   err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Aturan Successfully Get",
		"Data":    aturan,
	})
}
func UpdateAturan(c *gin.Context) {
	uid := c.Param("uid")

	var InputAturan struct {
		KodePenyakit             string `json:"kode_penyakit" binding:"required"`
		KodePertanyaan           string `json:"kode_pertanyaan" binding:"required"`
		MinValue                 int64  `json:"min_value"`
		IsMandatory              int64  `json:"is_mandatory"`
		TipeAturan               string `json:"tipe_aturan"`
		BerlakuUntukSemuaTingkat bool   `json:"berlaku_untuk_semua_tingkat"`
	}

	if err := c.ShouldBindJSON(&InputAturan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Silakan Ulangi Lagi Input Aturan",
			"Error":   err,
		})
		return
	}

	var aturan models.Aturan

	if err := database.DB.Where("aturan_uid = ?", uid).First(&aturan).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Data Aturan Not Found",
			"Error":   err,
		})
		return
	}

	// Snapshot data sebelum update (Fitur 3)
	dataSebelum := aturan

	if InputAturan.KodePenyakit != aturan.KodePenyakit {
		aturan.KodePenyakit = InputAturan.KodePenyakit
	}

	if InputAturan.KodePertanyaan != aturan.KodePertanyaan {
		aturan.KodePertanyaan = InputAturan.KodePertanyaan
	}

	if InputAturan.MinValue != aturan.MinValue {
		aturan.MinValue = InputAturan.MinValue
	}

	if InputAturan.IsMandatory != aturan.IsMandatory {
		aturan.IsMandatory = InputAturan.IsMandatory
	}

	tipe := InputAturan.TipeAturan
	if tipe == "" {
		tipe = "GEJALA_INTI"
	}
	aturan.TipeAturan = tipe
	aturan.BerlakuUntukSemuaTingkat = InputAturan.BerlakuUntukSemuaTingkat

	if err := aturan.UpdateAturan(uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Data Aturan Gagal Terupdate",
			"Error":   err,
		})
		return
	}

	// Audit Log (Fitur 3)
	utils.WriteAuditLog(c, "aturans", uid, "UPDATE", dataSebelum, aturan)

	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Aturan Successfully Updated",
		"Data":    aturan,
	})
}
func DeleteAturan(c *gin.Context) {
	uid := c.Param("uid")

	// Fetch data sebelum delete untuk audit log (Fitur 3)
	var dataSebelum models.Aturan
	if err := database.DB.Where("aturan_uid = ?", uid).First(&dataSebelum).Error; err == nil {
		utils.WriteAuditLog(c, "aturans", uid, "DELETE", dataSebelum, nil)
	}

	err := models.DeleteAturan(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Bad Request",
			"Message": "Data Aturan Not Found",
			"Error":   err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Aturan Successfully Deleted",
		"Data":    err,
	})
}
