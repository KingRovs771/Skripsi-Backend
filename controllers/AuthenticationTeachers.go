package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LoginTeachers(c *gin.Context) {
	var LoginTeachers struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&LoginTeachers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Input Invalid",
			"Error":   err.Error(),
		})
		return
	}

	var TeachersModel models.Teachers
	if err := database.DB.Where("email = ?", LoginTeachers.Email).First(&TeachersModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"Status":  "Error",
				"Message": "Akun Guru Tidak Ada, Silakan Hubungi Administrator Untuk Membuat Akun",
				"Error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Status Internal Server Error",
			"Error":   err.Error(),
		})
		return
	}
	if err := TeachersModel.ValidatePassword(LoginTeachers.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Email atau password salah",
			"Error":   err.Error(),
		})
		return
	}

	token, err := utils.GenerateJWT(TeachersModel.NIP, TeachersModel.Email, "Teachers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Message": "Login berhasil",
		"Token":   token,
		"User": gin.H{
			"NIP":          TeachersModel.NIP,
			"teachers_uid": TeachersModel.TeachersUID,
			"nama_lengkap": TeachersModel.NamaLengkap,
			"email":        TeachersModel.Email,
		},
	})
}

func GetProfileTeachers(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	NIP := claims.ID

	var TeachersProfile models.Teachers
	if err := database.DB.Select("teachers.n_ip, teachers.teachers_uid, teachers.role_uid,"+
		"teachers.nama_lengkap, teachers.phone, teachers.email,"+
		"teachers.alamat, teachers.created_at, teachers.update_at,"+
		"roles.role_name",
	).Joins("left join roles on roles.role_uid = teachers.role_uid").Where("teachers.n_ip = ?", NIP).First(&TeachersProfile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Profile Teachers not found",
			"Error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Teachers Profile Found",
		"Data": gin.H{
			"nip":          TeachersProfile.NIP,
			"teacher_uid":  TeachersProfile.TeachersUID,
			"role_uid":     TeachersProfile.RoleUID,
			"nama_lengkap": TeachersProfile.NamaLengkap,
			"phone":        TeachersProfile.Phone,
			"email":        TeachersProfile.Email,
			"alamat":       TeachersProfile.Alamat,
			"created_at":   TeachersProfile.CreatedAt,
			"updated_at":   TeachersProfile.UpdateAt,
		},
	})
}
