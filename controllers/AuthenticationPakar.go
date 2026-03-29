package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LoginPakar(c *gin.Context) {
	var LoginPakar struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Menggunakan ShouldBindJSON agar konsisten dengan Admin
	if err := c.ShouldBindJSON(&LoginPakar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Email and Password",
			"Error":   err.Error(),
		})
		return
	}

	var pakar models.Pakar

	// Pengecekan Database
	if err := database.DB.Where("email = ?", LoginPakar.Email).First(&pakar).Error; err != nil {
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

	// Validasi Password
	if err := pakar.ValidatePassword(LoginPakar.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Email atau password salah",
			"Error":   err.Error(),
		})
		return
	}

	// Generate JWT (Menggunakan NomorSIP sebagai ID sesuai kodingan asli kamu)
	token, err := utils.GenerateJWT(pakar.PakarUID, pakar.Email, "Pakar")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal Server Error",
			"Error":   err.Error(),
		})
		return
	}

	pakarCache := gin.H{
		"nomor_sip":    pakar.NomorSIP,
		"pakar_uid":    pakar.PakarUID,
		"nama_lengkap": pakar.NamaLengkap,
		"email":        pakar.Email,
		"role":         "Pakar",
	}

	// Marshal data ke JSON
	jsonData, _ := json.Marshal(pakarCache)

	tokenLifespanStr := os.Getenv("TOKEN_HOUR_LIFESPAN")
	tokenLifespan, _ := strconv.Atoi(tokenLifespanStr)
	if tokenLifespan == 0 {
		tokenLifespan = 1
	}

	// Simpan ke Redis (Menggunakan PakarUID sebagai Key)
	err = database.RDB.Set(database.Ctx, "profile:"+pakar.PakarUID, jsonData, time.Hour*time.Duration(tokenLifespan)).Err()
	if err != nil {
		fmt.Println("Gagal menyimpan cache ke Redis:", err)
	}

	// Response Sukses
	c.JSON(http.StatusOK, gin.H{
		"Status":  "OK",
		"Message": "Login Success",
		"Token":   token,
		"User": gin.H{
			"nomor_sip":    pakar.NomorSIP,
			"pakar_uid":    pakar.PakarUID,
			"nama_lengkap": pakar.NamaLengkap,
			"email":        pakar.Email,
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
	if err := database.DB.Select("pakars.nomor_s_ip, pakars.pakar_uid, pakars.role_uid,"+
		"pakars.nama_lengkap, pakars.phone, pakars.email,"+
		"pakars.alamat, pakars.created_at,"+
		"roles.role_name",
	).Joins("left join roles on roles.role_uid = pakars.role_uid").Where("pakars.nomor_s_ip = ?", NomorSIP).First(&PakarProfiles).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Profile Pakar not found",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Pakars Profile Found",
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
