package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ─── UpdateStudentStatus ──────────────────────────────────────────────────────
// PUT /api/admin/retention/students/:uid/status
// Body: { "status_akun": "AKTIF" | "LULUS" | "PINDAH" }

type UpdateStatusRequest struct {
	StatusAkun string `json:"status_akun" binding:"required"`
}

func UpdateStudentStatus(c *gin.Context) {
	uid := c.Param("uid")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.StatusAkun != "AKTIF" && req.StatusAkun != "LULUS" && req.StatusAkun != "PINDAH") {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Status akun tidak valid. Gunakan 'AKTIF', 'LULUS', atau 'PINDAH'",
		})
		return
	}

	var student models.Students
	if err := database.DB.Where("students_uid = ?", uid).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Siswa tidak ditemukan",
		})
		return
	}

	// Update status
	student.StatusAkun = req.StatusAkun
	if req.StatusAkun == "AKTIF" {
		student.TanggalNonaktif = nil
	} else {
		now := time.Now()
		student.TanggalNonaktif = &now
	}

	if err := database.DB.Save(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal memperbarui status siswa",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Status akun siswa berhasil diperbarui",
		"Data":    student,
	})
}

// ─── GetNearingDeletionStudents ──────────────────────────────────────────────
// GET /api/admin/retention/nearing-deletion
//
// Mengembalikan daftar siswa LULUS/PINDAH beserta info jumlah hari yang tersisa
// sebelum data mereka dianonimkan (retensi 2 tahun).

type NearingDeletionResponse struct {
	StudentsUID     string     `json:"students_uid"`
	NISN            string     `json:"nisn"`
	NamaLengkap     string     `json:"nama_lengkap"`
	Kelas           string     `json:"kelas"`
	NamaSekolah     string     `json:"nama_sekolah"`
	StatusAkun      string     `json:"status_akun"`
	TanggalNonaktif *time.Time `json:"tanggal_nonaktif"`
	DaysRemaining   int        `json:"days_remaining"`
}

func GetNearingDeletionStudents(c *gin.Context) {
	var list []models.Students
	// Cari semua siswa yang tidak aktif
	err := database.DB.Table("students").
		Select("students.*, sekolahs.nama_sekolah").
		Joins("left join sekolahs on sekolahs.npsn::text = students.npsn").
		Where("students.status_akun != 'AKTIF' AND students.tanggal_nonaktif IS NOT NULL").
		Order("students.tanggal_nonaktif ASC").
		Scan(&list).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil data retensi siswa",
		})
		return
	}

	retentionPeriod := 2 * 365 * 24 * time.Hour // 2 tahun

	result := make([]NearingDeletionResponse, 0, len(list))
	for _, student := range list {
		if student.TanggalNonaktif == nil {
			continue
		}

		deletionTime := student.TanggalNonaktif.Add(retentionPeriod)
		timeLeft := time.Until(deletionTime)
		daysLeft := int(math.Ceil(timeLeft.Hours() / 24))

		// Hanya tampilkan jika belum dianonimkan (atau tampilkan minus jika cron job belum jalan)
		result = append(result, NearingDeletionResponse{
			StudentsUID:     student.StudentsUID,
			NISN:            student.NISN,
			NamaLengkap:     student.NamaLengkap,
			Kelas:           student.Kelas,
			NamaSekolah:     student.NamaSekolah,
			StatusAkun:      student.StatusAkun,
			TanggalNonaktif: student.TanggalNonaktif,
			DaysRemaining:   daysLeft,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data":   result,
	})
}
