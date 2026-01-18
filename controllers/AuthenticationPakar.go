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

func LoginPakar(c *gin.Context) {
	var LoginPakar struct {
		Email    string `json:"email" binding:"required, email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBind(&LoginPakar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Email and Password",
			"Error":   err.Error(),
		})
		return
	}

	var PakarModels models.Pakar
	if err := database.DB.Where("email = ?", LoginPakar.Email).First(&PakarModels).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"Status":  "Error",
				"Message": "User Not Found",
				"Error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal Server Error",
			"Error":   err.Error(),
		})
		return
	}

	if err := PakarModels.ValidatePassword(LoginPakar.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Email dan Password Salah",
			"Error":   err.Error(),
		})
		return
	}

	token, err := utils.GenerateJWT(PakarModels.NomorSIP, PakarModels.Email, "Pakar")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal Server Error",
			"Error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Login Success",
		"Token":   token,
		"User": gin.H{
			"nomor_sip":    PakarModels.NomorSIP,
			"pakar_uid":    PakarModels.PakarUID,
			"nama_lengkap": PakarModels.NamaLengkap,
			"email":        PakarModels.Email,
		},
	})
}

func GetProfilePakar(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Invalid Token",
			"Error":   err.Error(),
		})
		return
	}
	NomorSIP := claims.ID

	var PakarProfiles models.Pakar
	if err := database.DB.Select("pakars.nomor_sip, pakars.pakar_uid, pakars.role_uid,"+
		"pakars.nama_lengkap, pakars.phone, pakars.email,"+
		"pakars.alamat, pakars.created_at, pakars.updated_at,"+
		"roles.role_name",
	).Joins("left join roles on roles.role_uid = teachers.role_uid").Where("pakars.nomor_sip = ?", NomorSIP).First(&PakarProfiles).Error; err != nil {
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
			"nip":          PakarProfiles.NomorSIP,
			"pakar_uid":    PakarProfiles.PakarUID,
			"role_uid":     PakarProfiles.RoleUID,
			"nama_lengkap": PakarProfiles.NamaLengkap,
			"phone":        PakarProfiles.Phone,
			"email":        PakarProfiles.Email,
			"alamat":       PakarProfiles.Alamat,
			"created_at":   PakarProfiles.CreatedAt,
			"updated_at":   PakarProfiles.UpdateAt,
		},
	})
}
