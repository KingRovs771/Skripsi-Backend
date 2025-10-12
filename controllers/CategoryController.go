package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryInput struct{
	NameCategory string `json:"name_category"`
	Deskripsi string `json:"deskripsi"`
}

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

func GetCategoryByUID(c *gin.Context){
	uid := c.Param("uid")
	var category models.Category

	if err := database.DB.Where("category_uid = ?", uid).First(category).Error; err !=nil{
		if err == gorm.ErrRecordNotFound{
			c.JSON(http.StatusNotFound, gin.H{
				"Status" : "Error",
				"Message" : "Category Not Found",
				"Error" : err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status" : "Error",
			"Message" : "Gagal Mendapatkan Data Category",
			"Error" : err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"Status" : "Success",
		"Message" : "Berhasil Mendapatkan Data Category",
		"Data": category,
	})
}

func UpdateCategory(c *gin.Context){
	uid := c.Param("uid")

	var category models.Category

	if err := database.DB.Where("category_uid = ?", uid).First(category).Error; err != nil{
		c.JSON(http.StatusNotFound, gin.H{
			"Status" : "Error",
			"Message" : "Data Category Tidak Ada",
			"Error" : err.Error(),
		})
		return
	}

	var inputUpdate CategoryInput
	if err := c.ShouldBindBodyWithJSON(&inputUpdate); err !=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status" : "Error",
			"Message" : "Invalid Input Data",
			"Error" : err.Error(),
		})
	}

	if err := database.DB.Model(&category).Updates(inputUpdate).Error;err !=nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status" : "Error",
			"Message" : "Failed Update Data Category",
			"Error" : err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status" : "Success",
		"Message" : "Data Berhasil di Update",
		"Data" : category,
	})
}

func DeleteCategory(c *gin.Context){
	uid := c.Param("uid")

	result := database.DB.Where("category_uid = ?", uid).Delete(&models.Category{})

	if result.Error != nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status" : "Error",
			"Message" : "Delete Category Gagal Silakan Coba lagi",
		})
		return
	}

	if result.RowsAffected == 0{
		c.JSON(http.StatusNotFound, gin.H{
			"Status" : "Error",
			"Message" : "Category Tidak Ada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status" : "Success",
		"Message" : "Category Berhasil di Hapus",
		"Data" : result,
	})
}