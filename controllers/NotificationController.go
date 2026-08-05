package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helper — buat notifikasi untuk siswa (dipanggil dari controller lain)
// ─────────────────────────────────────────────────────────────────────────────

func createStudentNotification(studentUID, notifType, title, message, relatedUID string) {
	notif := models.StudentNotification{
		StudentUID: studentUID,
		Type:       notifType,
		Title:      title,
		Message:    message,
		RelatedUID: relatedUID,
	}
	if err := database.DB.Create(&notif).Error; err != nil {
		log.Printf("[Notification] Gagal membuat notifikasi untuk student_uid=%s: %v", studentUID, err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SISWA — Ambil semua notifikasi (belum dibaca paling atas)
// GET /api/siswa/notifications
// ─────────────────────────────────────────────────────────────────────────────

func GetStudentNotifications(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	var notifs []models.StudentNotification
	if err := database.DB.
		Where("student_uid = ?", claims.ID).
		Order("is_read ASC, created_at DESC").
		Limit(50).
		Find(&notifs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat notifikasi"})
		return
	}

	var unreadCount int64
	database.DB.Model(&models.StudentNotification{}).
		Where("student_uid = ? AND is_read = false", claims.ID).
		Count(&unreadCount)

	c.JSON(http.StatusOK, gin.H{
		"Data":         notifs,
		"unread_count": unreadCount,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// SISWA — Tandai satu notifikasi sebagai sudah dibaca
// POST /api/siswa/notifications/:notif_uid/read
// ─────────────────────────────────────────────────────────────────────────────

func MarkNotificationRead(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	notifUID := c.Param("notif_uid")
	result := database.DB.Model(&models.StudentNotification{}).
		Where("notif_uid = ? AND student_uid = ?", notifUID, claims.ID).
		Update("is_read", true)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update notifikasi"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notifikasi tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notifikasi ditandai sudah dibaca"})
}

// ─────────────────────────────────────────────────────────────────────────────
// SISWA — Tandai SEMUA notifikasi sebagai sudah dibaca
// POST /api/siswa/notifications/read-all
// ─────────────────────────────────────────────────────────────────────────────

func MarkAllNotificationsRead(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}
	database.DB.Model(&models.StudentNotification{}).
		Where("student_uid = ? AND is_read = false", claims.ID).
		Update("is_read", true)
	c.JSON(http.StatusOK, gin.H{"message": "Semua notifikasi ditandai sudah dibaca"})
}

// ─────────────────────────────────────────────────────────────────────────────
// SISWA — Unread count saja (polling ringan dari layout)
// GET /api/siswa/notifications/unread-count
// ─────────────────────────────────────────────────────────────────────────────

func GetUnreadNotificationCount(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}
	var count int64
	database.DB.Model(&models.StudentNotification{}).
		Where("student_uid = ? AND is_read = false", claims.ID).
		Count(&count)
	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// ─────────────────────────────────────────────────────────────────────────────
// ADMIN — Statistik laporan bully per sekolah & per bulan
// GET /api/admin/bully-reports/stats?npsn=...&year=2026
// ─────────────────────────────────────────────────────────────────────────────

type BullyStatsBySchool struct {
	NPSN         string `json:"npsn"`
	NamaSekolah  string `json:"nama_sekolah"`
	Total        int64  `json:"total"`
	TotalBaru    int64  `json:"total_baru"`
	TotalSelesai int64  `json:"total_selesai"`
	TotalDarurat int64  `json:"total_darurat"`
}

type BullyStatsByMonth struct {
	Bulan   string `json:"bulan"` // "2026-08"
	Total   int64  `json:"total"`
	Baru    int64  `json:"baru"`
	Selesai int64  `json:"selesai"`
}

type BullyGlobalSummary struct {
	TotalAll     int64 `json:"total_all"`
	TotalBaru    int64 `json:"total_baru"`
	TotalDarurat int64 `json:"total_darurat"`
	TotalSelesai int64 `json:"total_selesai"`
}

func GetAdminBullyStats(c *gin.Context) {
	_, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	npsnFilter := c.Query("npsn")
	year       := c.DefaultQuery("year", "")

	// ── Builder kondisi dinamis ────────────────────────────────────────────
	baseWhere := "WHERE 1=1"
	args      := []interface{}{}
	if npsnFilter != "" {
		baseWhere += " AND br.npsn = ?"
		args = append(args, npsnFilter)
	}
	if year != "" {
		baseWhere += " AND EXTRACT(YEAR FROM br.created_at) = ?"
		args = append(args, year)
	}

	// ── Per Sekolah ────────────────────────────────────────────────────────
	var bySchool []BullyStatsBySchool
	schQuery := `
		SELECT
		  br.npsn,
		  COALESCE(s.nama_sekolah, br.npsn) AS nama_sekolah,
		  COUNT(*)                                                          AS total,
		  COUNT(*) FILTER (WHERE br.status = 'BARU')                       AS total_baru,
		  COUNT(*) FILTER (WHERE br.status = 'SELESAI')                    AS total_selesai,
		  COUNT(*) FILTER (WHERE br.tingkat_urgensi = 'Tinggi/Darurat')    AS total_darurat
		FROM bully_reports br
		LEFT JOIN sekolahs s ON s.npsn::text = br.npsn
		` + baseWhere + `
		GROUP BY br.npsn, s.nama_sekolah
		ORDER BY total DESC`

	if err := database.DB.Raw(schQuery, args...).Scan(&bySchool).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat statistik per sekolah"})
		return
	}

	// ── Per Bulan ──────────────────────────────────────────────────────────
	var byMonth []BullyStatsByMonth
	monthQuery := `
		SELECT
		  TO_CHAR(br.created_at, 'YYYY-MM')                           AS bulan,
		  COUNT(*)                                                     AS total,
		  COUNT(*) FILTER (WHERE br.status = 'BARU')                  AS baru,
		  COUNT(*) FILTER (WHERE br.status = 'SELESAI')               AS selesai
		FROM bully_reports br
		` + baseWhere + `
		GROUP BY bulan
		ORDER BY bulan DESC
		LIMIT 24`

	if err := database.DB.Raw(monthQuery, args...).Scan(&byMonth).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat statistik per bulan"})
		return
	}

	// ── Ringkasan Global ───────────────────────────────────────────────────
	// Rebuild where tanpa alias "br."
	globalWhere := "WHERE 1=1"
	globalArgs  := []interface{}{}
	if npsnFilter != "" {
		globalWhere += " AND npsn = ?"
		globalArgs = append(globalArgs, npsnFilter)
	}
	if year != "" {
		globalWhere += " AND EXTRACT(YEAR FROM created_at) = ?"
		globalArgs = append(globalArgs, year)
	}

	var summary BullyGlobalSummary
	database.DB.Raw(`
		SELECT
		  COUNT(*)                                                        AS total_all,
		  COUNT(*) FILTER (WHERE status = 'BARU')                        AS total_baru,
		  COUNT(*) FILTER (WHERE tingkat_urgensi = 'Tinggi/Darurat')     AS total_darurat,
		  COUNT(*) FILTER (WHERE status = 'SELESAI')                     AS total_selesai
		FROM bully_reports `+globalWhere, globalArgs...).Scan(&summary)

	c.JSON(http.StatusOK, gin.H{
		"Data": gin.H{
			"summary":   summary,
			"by_school": bySchool,
			"by_month":  byMonth,
		},
	})
}
