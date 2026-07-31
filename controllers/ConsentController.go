package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// currentConsentVersion mengembalikan versi kebijakan aktif dari env var.
// Default "v1" jika env var tidak di-set.
func currentConsentVersion() string {
	if v := os.Getenv("CONSENT_CURRENT_VERSION"); v != "" {
		return v
	}
	return "v1"
}

// ─── GetConsentStatus ─────────────────────────────────────────────────────────
// GET /api/siswa/consent/status
//
// Frontend memanggil ini saat halaman tes dimuat untuk menentukan apakah
// siswa perlu diarahkan ke halaman consent terlebih dahulu.

func GetConsentStatus(c *gin.Context) {
	studentUID, exists := c.Get("user_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Token tidak valid",
		})
		return
	}

	version := currentConsentVersion()

	var consent models.StudentConsent
	err := database.DB.
		Where("student_uid = ? AND consent_version = ? AND is_agreed = true", studentUID, version).
		First(&consent).Error

	hasConsented := err == nil

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data": gin.H{
			"has_consented":   hasConsented,
			"consent_version": version,
			"agreed_at":       consent.AgreedAt,
		},
	})
}

// ─── SubmitConsent ─────────────────────────────────────────────────────────────
// POST /api/siswa/consent/submit
// Body: { "is_agreed": true | false }
//
// Dicatat ke DB baik jika setuju maupun tidak — keduanya merupakan jejak audit
// yang diperlukan untuk compliance UU PDP.

type SubmitConsentRequest struct {
	IsAgreed bool `json:"is_agreed"`
}

func SubmitConsent(c *gin.Context) {
	studentUID, exists := c.Get("user_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Token tidak valid",
		})
		return
	}

	var req SubmitConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Format request tidak valid",
		})
		return
	}

	version := currentConsentVersion()
	now := time.Now()
	ipAddr := c.ClientIP()

	record := models.StudentConsent{
		ConsentUID:     uuid.New().String(),
		StudentUID:     studentUID.(string),
		ConsentVersion: version,
		IsAgreed:       req.IsAgreed,
		IPAddress:      ipAddr,
		CreatedAt:      now,
	}

	if req.IsAgreed {
		record.AgreedAt = &now
	}

	if err := database.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal menyimpan persetujuan",
		})
		return
	}

	if req.IsAgreed {
		c.JSON(http.StatusOK, gin.H{
			"Status":  "Success",
			"Message": "Persetujuan berhasil disimpan. Anda sekarang dapat memulai tes.",
			"Data":    gin.H{"is_agreed": true, "consent_version": version},
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Status":  "Success",
			"Message": "Penolakan Anda telah dicatat. Anda tidak dapat menggunakan fitur tes selama belum memberikan persetujuan.",
			"Data":    gin.H{"is_agreed": false},
		})
	}
}
