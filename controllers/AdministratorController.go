package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UpdateAdminInput struct {
	NamaLengkap string `json:"nama_lengkap"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Alamat      string `json:"alamat"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	RoleUID     string `json:"role_id"`
}

func LoginAdmin(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loginAdmin, err := models.FindUserByEmailAdministrator(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau password salah"})
		return
	}

	err = loginAdmin.ValidatePasswordAdministrator(input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau password salah"})
		return
	}
	jwt, err := utils.GenerateJWTAdmin(loginAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwt})
}

func GetProfileAdmin(c *gin.Context) {
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

func GetAllAdministrator(c *gin.Context) {
	var admins []models.Administrator

	if admins == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Data Not Found",
		})
		return
	}
	if err := database.DB.Find(&admins).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Tidak Dapat Mendapatkan Data Administrator",
			"Error":   err.Error(),
		})
		return
	}

	for i := range admins {
		admins[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Administrator",
		"Data":    admins,
	})
}

func CreateAdministrator(c *gin.Context) {
	var admin models.Administrator

	if err := c.ShouldBindJSON(&admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"Message": "HTTP Bad Request",
		})
		return
	}

	if err := admin.BeforeSaveAdministrator(database.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error":   err.Error(),
			"Message": "Status Internal Server Error",
		})
		return
	}
	saveAdmin, err := admin.SaveAdministrator()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error":   err.Error(),
			"Message": "Status Internal Server Error",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"Status":  "200 - ",
		"Message": "Administrator Berhasil Di Buat, Silakan check pada Halaman Utama Administrator",
		"Data":    saveAdmin,
	})
}

func GetAdministratorByUID(c *gin.Context) {
	var admin models.Administrator
	uid := c.Param("uid")

	if err := database.DB.Where("admin_uid = ?", uid).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"Message": "Administrator Not Found",
				"Error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"Message": "Database Error",
			"Error":   err.Error(),
		})
	}

	admin.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Administrator",
		"Data":    admin,
	})
}

func DeleteAdministrator(c *gin.Context) {
	uid := c.Param("uid")

	result := database.DB.Where("admin_uid = ?", uid).Delete(&models.Administrator{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Failed to Delete Administrator",
			"Error":   result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Not Found",
			"Message": "Administrator Not Found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Administrator Berhasil Di Hapus",
		"Data":    result,
	})

}

func UpdateAdministrator(c *gin.Context) {
	uid := c.Param("uid")

	var admin models.Administrator

	if err := database.DB.Where("admin_uid = ?", uid).First(&admin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Administrator Not Found",
		})
		return
	}

	var inputAdmin UpdateAdminInput
	if err := c.ShouldBindBodyWithJSON(&inputAdmin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Data",
			"Error":   err.Error(),
		})
		return
	}

	if inputAdmin.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(inputAdmin.Password), 12)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Status":  "Error",
				"Message": "Failed to hash password",
			})
			return
		}
		inputAdmin.Password = string(hashedPassword)
	}

	if err := database.DB.Model(&admin).Updates(inputAdmin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Failed to update administrator",
			"Error":   err.Error(),
		})
		return
	}

	admin.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Administrator updated successfully",
		"Data":    admin,
	})

}
