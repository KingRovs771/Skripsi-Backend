package controllers

import (
	"Skripsi-Backend/crypto"
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type StudentHistorySummary struct {
	NISN             string    `json:"nisn"`
	Nama             string    `json:"nama"`
	Kelas            string    `json:"kelas"`
	TotalTes         int64     `json:"total_tes"`
	TerakhirTes      time.Time `json:"terakhir_tes"`
	SkorPHQ9         int64     `json:"skor_phq9"`
	SkorGAD7         int64     `json:"skor_gad7"`
	DepresiPenyakit  string    `json:"depresi_penyakit"`
	CemasPenyakit    string    `json:"cemas_penyakit"`
}

type StudentInfo struct {
	NISN        string `json:"nisn"`
	Nama        string `json:"nama"`
	Kelas       string `json:"kelas"`
	Email       string `json:"email"`
	StudentsUID string `json:"students_uid"`
}

type TestResult struct {
	ID                  int       `gorm:"column:id" json:"id"`
	SessionUID          string    `gorm:"column:session_uid" json:"session_uid"`
	NamaTes             string    `gorm:"column:nama_tes" json:"nama_tes"`
	Skor                int64     `gorm:"column:skor" json:"skor"`
	SkorPHQ9            int64     `gorm:"column:skor_phq9" json:"skor_phq9"`
	SkorGAD7            int64     `gorm:"column:skor_gad7" json:"skor_gad7"`
	Kategori            string    `gorm:"column:kategori" json:"kategori"`
	DepresiPenyakit     string    `gorm:"column:depresi_penyakit" json:"depresi_penyakit"`
	CemasPenyakit       string    `gorm:"column:cemas_penyakit" json:"cemas_penyakit"`
	Tanggal             time.Time `gorm:"column:tanggal" json:"tanggal"`
	Rekomendasi         string    `gorm:"column:rekomendasi" json:"rekomendasi"`
	CeritaSiswa         string    `gorm:"column:cerita_siswa" json:"cerita_siswa"`
	IsVisibleToStudent  bool      `gorm:"column:is_visible_to_student" json:"is_visible_to_student"`
	ReviewedByGurubk    bool      `gorm:"column:reviewed_by_gurubk" json:"reviewed_by_gurubk"`
	NNDepresiConfidence float64   `gorm:"column:nn_depresi_confidence" json:"nn_depresi_confidence"`
	NNCemasConfidence   float64   `gorm:"column:nn_cemas_confidence" json:"nn_cemas_confidence"`
	StatusValidasiDepresi string  `gorm:"column:status_validasi_depresi" json:"status_validasi_depresi"`
	StatusValidasiCemas   string  `gorm:"column:status_validasi_cemas" json:"status_validasi_cemas"`
	DepresiSaran        string    `gorm:"column:depresi_saran" json:"depresi_saran"`
	CemasSaran          string    `gorm:"column:cemas_saran" json:"cemas_saran"`
}

type ReviewRequest struct {
	IsVisibleToStudent bool   `json:"is_visible_to_student"`
	ReviewedByGurubk   bool   `json:"reviewed_by_gurubk"`
	Rekomendasi        string `json:"rekomendasi"`
}

func GetGurubkHistory(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}

	var teacherNPSN string
	if err := database.DB.Table("teachers").Where("n_ip = ?", claims.ID).Select("npsn").Scan(&teacherNPSN).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teacher profile"})
		return
	}

	var results []StudentHistorySummary
	query := `
		SELECT
		  s.nisn,
		  s.nama_lengkap AS nama,
		  s.kelas,
		  COUNT(ts.test_session_id) AS total_tes,
		  MAX(ts.created_at) AS terakhir_tes,
		  last_ts.total_scorephq9 AS skor_phq9,
		  last_ts.total_scoregad7 AS skor_gad7,
		  p_depresi.nama_penyakit AS depresi_penyakit,
		  p_cemas.nama_penyakit AS cemas_penyakit
		FROM students s
		JOIN test_sessions ts ON s.students_uid = ts.user_uid AND ts.status = 'SELESAI'
		JOIN LATERAL (
		  SELECT ts2.total_scorephq9, ts2.total_scoregad7, ts2.test_session_id
		  FROM test_sessions ts2
		  WHERE ts2.user_uid = s.students_uid AND ts2.status = 'SELESAI'
		  ORDER BY ts2.created_at DESC
		  LIMIT 1
		) last_ts ON TRUE
		JOIN hasil_diagnoses hd ON hd.session_test_uid = last_ts.test_session_id
		LEFT JOIN penyakits p_depresi ON hd.final_depresi_penyakit = p_depresi.kode_penyakit
		LEFT JOIN penyakits p_cemas ON hd.final_cemas_penyakit = p_cemas.kode_penyakit
		WHERE s.npsn = ?
		GROUP BY s.nisn, s.nama_lengkap, s.kelas, last_ts.total_scorephq9, last_ts.total_scoregad7, p_depresi.nama_penyakit, p_cemas.nama_penyakit
		ORDER BY MAX(ts.created_at) DESC
	`
	if err := database.DB.Raw(query, teacherNPSN).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Data": results,
	})
}

func GetGurubkHistoryDetail(c *gin.Context) {
	nisn := c.Param("nisn")
	var studentInfo StudentInfo
	if err := database.DB.Table("students").Where("nisn = ?", nisn).Select("nisn, nama_lengkap as nama, kelas, email, students_uid").Scan(&studentInfo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	var results []TestResult
	query := `
		SELECT hd.result_id as id, ts.test_session_id as session_uid, 
		       'Diagnosis Kesehatan Mental' as nama_tes,
		       (ts.total_scorephq9 + ts.total_scoregad7) as skor,
		       ts.total_scorephq9 as skor_phq9,
		       ts.total_scoregad7 as skor_gad7,
		       CONCAT(p_depresi.nama_penyakit, ' & ', p_cemas.nama_penyakit) as kategori,
		       p_depresi.nama_penyakit as depresi_penyakit,
		       p_cemas.nama_penyakit as cemas_penyakit,
		       ts.created_at as tanggal,
		       hd.rekomendasi,
		       sf.cerita_siswa,
		       hd.is_visible_to_student, hd.reviewed_by_gurubk,
		       hd.nn_depresi_confidence, hd.nn_cemas_confidence,
		       hd.status_validasi_depresi, hd.status_validasi_cemas,
		       COALESCE(p_depresi.saran_penanganan, '') as depresi_saran,
		       COALESCE(p_cemas.saran_penanganan, '') as cemas_saran
		FROM students s
		JOIN test_sessions ts ON s.students_uid = ts.user_uid
		JOIN hasil_diagnoses hd ON ts.test_session_id = hd.session_test_uid
		LEFT JOIN student_feedbacks sf ON ts.test_session_id = sf.test_session_uid
		LEFT JOIN penyakits p_depresi ON hd.final_depresi_penyakit = p_depresi.kode_penyakit
		LEFT JOIN penyakits p_cemas ON hd.final_cemas_penyakit = p_cemas.kode_penyakit
		WHERE s.nisn = ? AND ts.status = 'SELESAI'
		ORDER BY ts.created_at DESC
	`
	if err := database.DB.Raw(query, nisn).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Dekripsi cerita_siswa dan rekomendasi karena dibaca via raw SQL query scan (hooks AfterFind tidak trigger)
	for i := range results {
		if decrypted, err := crypto.DecryptField(results[i].CeritaSiswa); err == nil {
			results[i].CeritaSiswa = decrypted
		}
		if decrypted, err := crypto.DecryptField(results[i].Rekomendasi); err == nil {
			results[i].Rekomendasi = decrypted
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"Data": gin.H{
			"student": studentInfo,
			"results": results,
		},
	})
}

func UpdateHistoryReview(c *gin.Context) {
	resultId := c.Param("id")
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	// Enkripsi rekomendasi secara manual karena updates dengan map tidak men-trigger GORM BeforeSave/BeforeUpdate hooks
	encryptedRekomendasi, err := crypto.EncryptField(req.Rekomendasi)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi rekomendasi"})
		return
	}

	if err := database.DB.Model(&models.HasilDiagnosis{}).Where("result_id = ?", resultId).Updates(map[string]interface{}{
		"is_visible_to_student": req.IsVisibleToStudent,
		"reviewed_by_gurubk":    req.ReviewedByGurubk,
		"rekomendasi":           encryptedRekomendasi,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status berhasil diupdate"})
}

func GetStudentHistory(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}

	studentUID := claims.ID

	var results []TestResult
	query := `
		SELECT hd.result_id as id, ts.test_session_id as session_uid, 
		       'Diagnosis Kesehatan Mental' as nama_tes,
		       (ts.total_scorephq9 + ts.total_scoregad7) as skor,
		       ts.total_scorephq9 as skor_phq9,
		       ts.total_scoregad7 as skor_gad7,
		       CONCAT(p_depresi.nama_penyakit, ' & ', p_cemas.nama_penyakit) as kategori,
		       p_depresi.nama_penyakit as depresi_penyakit,
		       p_cemas.nama_penyakit as cemas_penyakit,
		       ts.created_at as tanggal,
		       hd.rekomendasi,
		       sf.cerita_siswa,
		       hd.is_visible_to_student, hd.reviewed_by_gurubk,
		       hd.nn_depresi_confidence, hd.nn_cemas_confidence,
		       hd.status_validasi_depresi, hd.status_validasi_cemas,
		       COALESCE(p_depresi.saran_penanganan, '') as depresi_saran,
		       COALESCE(p_cemas.saran_penanganan, '') as cemas_saran
		FROM test_sessions ts
		JOIN hasil_diagnoses hd ON ts.test_session_id = hd.session_test_uid
		LEFT JOIN student_feedbacks sf ON ts.test_session_id = sf.test_session_uid
		LEFT JOIN penyakits p_depresi ON hd.final_depresi_penyakit = p_depresi.kode_penyakit
		LEFT JOIN penyakits p_cemas ON hd.final_cemas_penyakit = p_cemas.kode_penyakit
		WHERE ts.user_uid = ? AND ts.status = 'SELESAI'
		ORDER BY ts.created_at DESC
	`
	if err := database.DB.Raw(query, studentUID).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Dekripsi cerita_siswa dan rekomendasi untuk siswa
	for i := range results {
		if decrypted, err := crypto.DecryptField(results[i].CeritaSiswa); err == nil {
			results[i].CeritaSiswa = decrypted
		}
		if decrypted, err := crypto.DecryptField(results[i].Rekomendasi); err == nil {
			results[i].Rekomendasi = decrypted
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"Data": results,
	})
}
