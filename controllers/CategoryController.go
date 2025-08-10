package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateCategory(c *gin.Context) {
	var input struct {
		Name        string `json:"name_category" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := models.Category{
		NameCategory: input.Name,
		Description:  input.Description,
	}

	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan kategori"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Kategori berhasil dibuat", "data": category})
}

func GetAllCategories(c *gin.Context) {
	var categories []models.Category
	if err := database.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kategori"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}
