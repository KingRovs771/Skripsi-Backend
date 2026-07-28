package controllers

import (
	"Skripsi-Backend/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ─── Shared structs ───────────────────────────────────────────────────────────

type SeverityCount struct {
	Normal int64 `json:"normal"`
	Ringan int64 `json:"ringan"`
	Sedang int64 `json:"sedang"`
	Berat  int64 `json:"berat"`
}

// ─── GetSchoolMonitoring ─────────────────────────────────────────────────────
// GET /api/admin/monitoring/sekolah?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
//
// Mengagregasi seluruh hasil diagnosis (termasuk yang belum divalidasi GurBK)
// dikelompokkan per sekolah. Severity ditentukan berdasarkan akhiran kode
// penyakit: x01=normal, x02=ringan, x03=sedang, x04=berat.

type SchoolMonitoringRow struct {
	NPSN           int64  `gorm:"column:npsn"            json:"npsn"`
	NamaSekolah    string `gorm:"column:nama_sekolah"    json:"nama_sekolah"`
	TotalDiagnoses int64  `gorm:"column:total_diagnoses" json:"total_diagnoses"`
	DepresiNormal  int64  `gorm:"column:depresi_normal"  json:"-"`
	DepresiRingan  int64  `gorm:"column:depresi_ringan"  json:"-"`
	DepresiSedang  int64  `gorm:"column:depresi_sedang"  json:"-"`
	DepresiBerat   int64  `gorm:"column:depresi_berat"   json:"-"`
	CemasNormal    int64  `gorm:"column:cemas_normal"    json:"-"`
	CemasRingan    int64  `gorm:"column:cemas_ringan"    json:"-"`
	CemasSedang    int64  `gorm:"column:cemas_sedang"    json:"-"`
	CemasBerat     int64  `gorm:"column:cemas_berat"     json:"-"`
	UrgentCount    int64  `gorm:"column:urgent_count"    json:"urgent_count"`
}

type SchoolMonitoringResponse struct {
	NPSN           int64         `json:"npsn"`
	NamaSekolah    string        `json:"nama_sekolah"`
	TotalDiagnoses int64         `json:"total_diagnoses"`
	Depresi        SeverityCount `json:"depresi"`
	Cemas          SeverityCount `json:"cemas"`
	UrgentCount    int64         `json:"urgent_count"`
}

func GetSchoolMonitoring(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := `
		SELECT
			sk.npsn,
			sk.nama_sekolah,
			COUNT(hd.result_id)                                                               AS total_diagnoses,
			COUNT(CASE WHEN hd.final_depresi_penyakit LIKE '%01' THEN 1 END)                 AS depresi_normal,
			COUNT(CASE WHEN hd.final_depresi_penyakit LIKE '%02' THEN 1 END)                 AS depresi_ringan,
			COUNT(CASE WHEN hd.final_depresi_penyakit LIKE '%03' THEN 1 END)                 AS depresi_sedang,
			COUNT(CASE WHEN hd.final_depresi_penyakit LIKE '%04' THEN 1 END)                 AS depresi_berat,
			COUNT(CASE WHEN hd.final_cemas_penyakit   LIKE '%01' THEN 1 END)                 AS cemas_normal,
			COUNT(CASE WHEN hd.final_cemas_penyakit   LIKE '%02' THEN 1 END)                 AS cemas_ringan,
			COUNT(CASE WHEN hd.final_cemas_penyakit   LIKE '%03' THEN 1 END)                 AS cemas_sedang,
			COUNT(CASE WHEN hd.final_cemas_penyakit   LIKE '%04' THEN 1 END)                 AS cemas_berat,
			COUNT(CASE WHEN hd.status_validasi_depresi = 'URGENT_INTERVENTION' THEN 1 END)   AS urgent_count
		FROM sekolahs sk
		LEFT JOIN students        s  ON sk.npsn::text = s.npsn
		LEFT JOIN test_sessions   ts ON s.students_uid = ts.user_uid AND ts.status = 'SELESAI'
		LEFT JOIN hasil_diagnoses hd ON ts.test_session_id = hd.session_test_uid
	`

	var args []interface{}
	if startDate != "" && endDate != "" {
		start, errS := time.Parse("2006-01-02", startDate)
		end, errE := time.Parse("2006-01-02", endDate)
		if errS == nil && errE == nil {
			end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query += " WHERE ts.created_at >= ? AND ts.created_at <= ?"
			args = append(args, start, end)
		}
	} else if startDate != "" {
		start, errS := time.Parse("2006-01-02", startDate)
		if errS == nil {
			query += " WHERE ts.created_at >= ?"
			args = append(args, start)
		}
	} else if endDate != "" {
		end, errE := time.Parse("2006-01-02", endDate)
		if errE == nil {
			end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query += " WHERE ts.created_at <= ?"
			args = append(args, end)
		}
	}

	query += " GROUP BY sk.npsn, sk.nama_sekolah ORDER BY sk.nama_sekolah"

	var rows []SchoolMonitoringRow
	if err := database.DB.Raw(query, args...).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil data monitoring sekolah",
			"Error":   err.Error(),
		})
		return
	}

	result := make([]SchoolMonitoringResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, SchoolMonitoringResponse{
			NPSN:           r.NPSN,
			NamaSekolah:    r.NamaSekolah,
			TotalDiagnoses: r.TotalDiagnoses,
			Depresi: SeverityCount{
				Normal: r.DepresiNormal,
				Ringan: r.DepresiRingan,
				Sedang: r.DepresiSedang,
				Berat:  r.DepresiBerat,
			},
			Cemas: SeverityCount{
				Normal: r.CemasNormal,
				Ringan: r.CemasRingan,
				Sedang: r.CemasSedang,
				Berat:  r.CemasBerat,
			},
			UrgentCount: r.UrgentCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Data monitoring sekolah berhasil diambil",
		"Data":    result,
	})
}

// ─── GetSchoolStudents ────────────────────────────────────────────────────────
// GET /api/admin/monitoring/sekolah/:npsn/siswa
//
// Daftar siswa per sekolah dengan ringkasan diagnosis terbaru.
// Privasi: cerita_siswa TIDAK disertakan.

type SchoolStudentRow struct {
	StudentsUID           string     `gorm:"column:students_uid"            json:"students_uid"`
	NamaLengkap           string     `gorm:"column:nama_lengkap"            json:"nama_lengkap"`
	NISN                  string     `gorm:"column:nisn"                    json:"nisn"`
	Kelas                 string     `gorm:"column:kelas"                   json:"kelas"`
	LastTestDate          *time.Time `gorm:"column:last_test_date"          json:"last_test_date"`
	TotalScorePHQ9        *int64     `gorm:"column:total_scorephq9"         json:"total_scorephq9"`
	TotalScoreGAD7        *int64     `gorm:"column:total_scoregad7"         json:"total_scoregad7"`
	KategoriDepresi       string     `gorm:"column:kategori_depresi"        json:"kategori_depresi"`
	KategoriCemas         string     `gorm:"column:kategori_cemas"          json:"kategori_cemas"`
	StatusValidasiDepresi string     `gorm:"column:status_validasi_depresi" json:"status_validasi_depresi"`
	StatusValidasiCemas   string     `gorm:"column:status_validasi_cemas"   json:"status_validasi_cemas"`
	IsUrgent              bool       `gorm:"column:is_urgent"               json:"is_urgent"`
}

func GetSchoolStudents(c *gin.Context) {
	npsn := c.Param("npsn")
	if npsn == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "Error", "Message": "NPSN tidak boleh kosong"})
		return
	}

	var namaSekolah string
	database.DB.Table("sekolahs").Where("npsn::text = ?", npsn).Select("nama_sekolah").Scan(&namaSekolah)
	if namaSekolah == "" {
		c.JSON(http.StatusNotFound, gin.H{"Status": "Error", "Message": "Sekolah tidak ditemukan"})
		return
	}

	query := `
		SELECT DISTINCT ON (s.students_uid)
			s.students_uid,
			s.nama_lengkap,
			s.nisn,
			s.kelas,
			ts.created_at                              AS last_test_date,
			ts.total_scorephq9,
			ts.total_scoregad7,
			COALESCE(p_dep.nama_penyakit, '')          AS kategori_depresi,
			COALESCE(p_cem.nama_penyakit, '')          AS kategori_cemas,
			COALESCE(hd.status_validasi_depresi, '')   AS status_validasi_depresi,
			COALESCE(hd.status_validasi_cemas,   '')   AS status_validasi_cemas,
			(hd.status_validasi_depresi = 'URGENT_INTERVENTION') AS is_urgent
		FROM students s
		LEFT JOIN test_sessions   ts    ON s.students_uid = ts.user_uid AND ts.status = 'SELESAI'
		LEFT JOIN hasil_diagnoses hd    ON ts.test_session_id = hd.session_test_uid
		LEFT JOIN penyakits       p_dep ON hd.final_depresi_penyakit = p_dep.kode_penyakit
		LEFT JOIN penyakits       p_cem ON hd.final_cemas_penyakit   = p_cem.kode_penyakit
		WHERE s.npsn = ?
		ORDER BY s.students_uid, ts.created_at DESC NULLS LAST
	`

	var students []SchoolStudentRow
	if err := database.DB.Raw(query, npsn).Scan(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil data siswa",
			"Error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Data siswa berhasil diambil",
		"Data": gin.H{
			"npsn":         npsn,
			"nama_sekolah": namaSekolah,
			"siswa":        students,
		},
	})
}

// ─── GetStudentHistory ────────────────────────────────────────────────────────
// GET /api/admin/monitoring/siswa/:students_uid/riwayat
//
// Seluruh riwayat tes satu siswa, terbaru lebih dulu.
// Privasi: cerita_siswa TIDAK diambil dari DB (bukan sekadar disembunyikan di FE).

type StudentHistoryRow struct {
	TestSessionID         string    `gorm:"column:test_session_id"         json:"test_session_id"`
	CreatedAt             time.Time `gorm:"column:created_at"              json:"created_at"`
	TotalScorePHQ9        int64     `gorm:"column:total_scorephq9"         json:"total_scorephq9"`
	TotalScoreGAD7        int64     `gorm:"column:total_scoregad7"         json:"total_scoregad7"`
	KategoriDepresi       string    `gorm:"column:kategori_depresi"        json:"kategori_depresi"`
	KategoriCemas         string    `gorm:"column:kategori_cemas"          json:"kategori_cemas"`
	FinalDepresiKode      string    `gorm:"column:final_depresi_penyakit"  json:"final_depresi_kode"`
	FinalCemasKode        string    `gorm:"column:final_cemas_penyakit"    json:"final_cemas_kode"`
	StatusValidasiDepresi string    `gorm:"column:status_validasi_depresi" json:"status_validasi_depresi"`
	StatusValidasiCemas   string    `gorm:"column:status_validasi_cemas"   json:"status_validasi_cemas"`
	NNDepresiConfidence   float64   `gorm:"column:nn_depresi_confidence"   json:"nn_depresi_confidence"`
	NNCemasConfidence     float64   `gorm:"column:nn_cemas_confidence"     json:"nn_cemas_confidence"`
	ReviewedByGurubk      bool      `gorm:"column:reviewed_by_gurubk"      json:"reviewed_by_gurubk"`
	// cerita_siswa sengaja TIDAK diambil — privasi Admin (UU PDP No.27/2022)
}

type StudentInfoForAdmin struct {
	StudentsUID string `gorm:"column:students_uid" json:"students_uid"`
	NamaLengkap string `gorm:"column:nama_lengkap" json:"nama_lengkap"`
	NISN        string `gorm:"column:nisn"         json:"nisn"`
	Kelas       string `gorm:"column:kelas"        json:"kelas"`
	NamaSekolah string `gorm:"column:nama_sekolah" json:"nama_sekolah"`
	NPSN        string `gorm:"column:npsn"         json:"npsn"`
}

func GetAdminStudentHistory(c *gin.Context) {
	studentsUID := c.Param("students_uid")
	if studentsUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "Error", "Message": "students_uid tidak boleh kosong"})
		return
	}

	var studentInfo StudentInfoForAdmin
	infoQuery := `
		SELECT s.students_uid, s.nama_lengkap, s.nisn, s.kelas,
		       COALESCE(sk.nama_sekolah, '') AS nama_sekolah,
		       s.npsn
		FROM students s
		LEFT JOIN sekolahs sk ON sk.npsn::text = s.npsn
		WHERE s.students_uid = ?
		LIMIT 1
	`
	if err := database.DB.Raw(infoQuery, studentsUID).Scan(&studentInfo).Error; err != nil || studentInfo.StudentsUID == "" {
		c.JSON(http.StatusNotFound, gin.H{"Status": "Error", "Message": "Siswa tidak ditemukan"})
		return
	}

	historyQuery := `
		SELECT
			ts.test_session_id,
			ts.created_at,
			ts.total_scorephq9,
			ts.total_scoregad7,
			COALESCE(p_dep.nama_penyakit, hd.final_depresi_penyakit, '') AS kategori_depresi,
			COALESCE(p_cem.nama_penyakit, hd.final_cemas_penyakit,   '') AS kategori_cemas,
			COALESCE(hd.final_depresi_penyakit, '')                       AS final_depresi_penyakit,
			COALESCE(hd.final_cemas_penyakit,   '')                       AS final_cemas_penyakit,
			COALESCE(hd.status_validasi_depresi, '')                      AS status_validasi_depresi,
			COALESCE(hd.status_validasi_cemas,   '')                      AS status_validasi_cemas,
			COALESCE(hd.nn_depresi_confidence, 0)                         AS nn_depresi_confidence,
			COALESCE(hd.nn_cemas_confidence,   0)                         AS nn_cemas_confidence,
			COALESCE(hd.reviewed_by_gurubk, false)                        AS reviewed_by_gurubk
		FROM test_sessions ts
		JOIN hasil_diagnoses hd   ON ts.test_session_id = hd.session_test_uid
		LEFT JOIN penyakits p_dep ON hd.final_depresi_penyakit = p_dep.kode_penyakit
		LEFT JOIN penyakits p_cem ON hd.final_cemas_penyakit   = p_cem.kode_penyakit
		WHERE ts.user_uid = ? AND ts.status = 'SELESAI'
		ORDER BY ts.created_at DESC
	`

	var history []StudentHistoryRow
	if err := database.DB.Raw(historyQuery, studentsUID).Scan(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil riwayat tes siswa",
			"Error":   err.Error(),
		})
		return
	}

	hasUrgent := false
	for _, h := range history {
		if h.StatusValidasiDepresi == "URGENT_INTERVENTION" {
			hasUrgent = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Riwayat tes siswa berhasil diambil",
		"Data": gin.H{
			"siswa":      studentInfo,
			"riwayat":    history,
			"has_urgent": hasUrgent,
		},
	})
}

