package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SekolahBinaanResponse struct {
	NPSN            int64  `json:"npsn"`
	SekolahUID      string `json:"sekolah_uid"`
	NamaSekolah     string `json:"nama_sekolah"`
	Jenjang         string `json:"jenjang"`
	AlamatSekolah   string `json:"alamat_sekolah"`
	TotalSiswa      int64  `json:"total_siswa"`
	TotalTesSelesai int64  `json:"total_tes_selesai"`
	ButuhPerhatian  int64  `json:"butuh_perhatian"`
}

func GetPakarSekolahBinaan(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if ID is pakar_uid or email/SIP. Let's resolve PakarUID.
	var pakarUIDResolved string
	if claims.UserType == "Pakar" {
		// Try to find the pakar record to get pakar_uid.
		// claims.ID can be PakarUID, NomorSIP or Email depending on how they login.
		// Let's check both pakar_uid and email.
		var p struct {
			PakarUID string
		}
		err := database.DB.Table("pakars").
			Select("pakar_uid").
			Where("pakar_uid = ? OR email = ?", claims.ID, claims.Email).
			First(&p).Error
		if err == nil {
			pakarUIDResolved = p.PakarUID
		} else {
			pakarUIDResolved = claims.ID
		}
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya Pakar yang dapat melihat sekolah binaan"})
		return
	}

	var results []SekolahBinaanResponse
	query := `
		SELECT 
			s.npsn,
			s.sekolah_uid,
			s.nama_sekolah,
			s.jenjang,
			s.alamat_sekolah,
			COALESCE((SELECT COUNT(*) FROM students stu WHERE stu.npsn::text = s.npsn::text), 0) as total_siswa,
			COALESCE((SELECT COUNT(*) FROM hasil_diagnoses hd 
			 JOIN test_sessions ts ON hd.session_test_uid = ts.test_session_id
			 JOIN students stu ON ts.user_uid = stu.students_uid
			 WHERE stu.npsn::text = s.npsn::text), 0) as total_tes_selesai,
			COALESCE((SELECT COUNT(*) FROM hasil_diagnoses hd 
			 JOIN test_sessions ts ON hd.session_test_uid = ts.test_session_id
			 JOIN students stu ON ts.user_uid = stu.students_uid
			 WHERE stu.npsn::text = s.npsn::text AND (hd.nn_depresi_confidence > 75 OR hd.nn_cemas_confidence > 75)), 0) as butuh_perhatian
		FROM sekolahs s
		WHERE s.pakar_uid = ?
	`

	if err := database.DB.Raw(query, pakarUIDResolved).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses data sekolah binaan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil memuat daftar sekolah binaan",
		"Data":    results,
	})
}
