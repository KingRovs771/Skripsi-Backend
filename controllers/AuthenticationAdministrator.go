package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func LoginAdministrator(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Bad Request",
			"Error":   err.Error(),
		})
		return
	}

	var admin models.Administrator

	if err := database.DB.Where("email = ?", input.Email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Status":  "Error",
				"Message": "Email atau password salah",
				"Error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Database error",
			"Error":   err.Error(),
		})
		return
	}

	if err := admin.ValidatePasswordAdministrator(input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Email atau password salah",
			"Error":   err.Error(),
		})
		return
	}

	token, err := utils.GenerateJWT(admin.AdminUID, admin.Email, "Administrator")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message": "Login berhasil",
		"Token":   token,
		"User": gin.H{
			"admin_id":     admin.AdminId,
			"admin_uid":    admin.AdminUID,
			"nama_lengkap": admin.NamaLengkap,
			"email":        admin.Email,
		},
	})
}
func GetProfileAdministrator(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	adminUID := claims.ID

	var admin models.Administrator
	if err := database.DB.Select("administrators.admin_id, administrators.admin_uid, administrators.role_uid,"+
		"administrators.nama_lengkap, administrators.phone, administrators.email,"+
		"administrators.alamat, administrators.created_at, administrators.updated_at,"+
		"roles.role_name",
	).Joins("left join roles on roles.role_uid = administrators.role_uid").Where("administrators.admin_uid = ?", adminUID).First(&admin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Profile administrator not found",
			"Error":   err.Error(),
		})
		return
	}

	// Kirim data profil tanpa password
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Administrator",
		"Data": gin.H{
			"admin_id":     admin.AdminId,
			"admin_uid":    admin.AdminUID,
			"role_uid":     admin.RoleUID,
			"nama_lengkap": admin.NamaLengkap,
			"phone":        admin.Phone,
			"email":        admin.Email,
			"alamat":       admin.Alamat,
			"created_at":   admin.CreatedAt,
			"updated_at":   admin.UpdateAt,
		},
	})
}
