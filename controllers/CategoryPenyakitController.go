package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SaveCategoryPenyakit(c *gin.Context) {
	var inputCategoryInput struct {
		NamaCategory string `form:"nama_category"`
		KodeCategory string `form:"kode_category"`
		Deskripsi    string `form:"deskripsi"`
	}

	if err := c.ShouldBindJSON(&inputCategoryInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Error Bad Request",
			"Error":   err.Error(),
		})
		return
	}

	cp := models.CategoryPenyakit{
		NamaCategory: inputCategoryInput.NamaCategory,
		KodeCategory: inputCategoryInput.KodeCategory,
		Deskripsi:    inputCategoryInput.Deskripsi,
	}
	hasil, err := cp.SaveCategoryPenyakits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Error Save Category Penyakit",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"Status":  "Created",
		"Message": "Save Category Penyakit",
		"Data":    hasil,
	})
}

func GetCategoryPenyakits(c *gin.Context) {
	cpList, err := models.GetAllCategoryPenyakits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Error Get All Category Penyakits",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Get All Category Penyakits",
		"Data":    cpList,
	})
}

func GetCategoryPenyakitsByUID(c *gin.Context) {
	uid := c.Param("uid")

	cp, err := models.GetCategoryPenyakitByUID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Error Get All Category Penyakit",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Get All Category Penyakits",
		"Data":    cp,
	})
}

func UpdateCategoryPenyakit(c *gin.Context) {
	uid := c.Param("uid")

	var inputCategoryInput struct {
		NamaCategory string `form:"nama_category"`
		KodeCategory string `form:"kode_category"`
		Deskripsi    string `form:"deskripsi"`
	}

	if err := c.ShouldBindJSON(
		&inputCategoryInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Error Bad Request",
			"Error":   err.Error(),
		})
		return
	}
	var cp models.CategoryPenyakit

	if err := database.DB.Where("category_penyakit_uid = ?", uid).First(&cp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Error Update Category Penyakit",
			"Error":   err.Error(),
		})
		return
	}
	if inputCategoryInput.NamaCategory != "" {
		cp.NamaCategory = inputCategoryInput.NamaCategory
	}
	if inputCategoryInput.KodeCategory != "" {
		cp.KodeCategory = inputCategoryInput.KodeCategory
	}
	if inputCategoryInput.Deskripsi != "" {
		cp.Deskripsi = inputCategoryInput.Deskripsi
	}

	if err := cp.UpdateCategoryPenyakits(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Error Update Category Penyakit",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Update Category Penyakit",
		"Data":    cp,
	})
}

func DeleteCategoryPenyakit(c *gin.Context) {
	uid := c.Param("uid")
	cp, err := models.GetCategoryPenyakitByUID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Error Delete Category Penyakit",
			"Error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Delete Category Penyakit",
		"Data":    cp,
	})
}
