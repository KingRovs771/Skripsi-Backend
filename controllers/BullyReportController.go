package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────────────────────────────────
// Request / Response Types
// ─────────────────────────────────────────────────────────────────────────────

type SubmitBullyReportRequest struct {
	JenisBully        string `json:"jenis_bully"        binding:"required"`
	NamaTerlapor      string `json:"nama_terlapor"`
	KelasTerlapor     string `json:"kelas_terlapor"`
	DeskripsiKejadian string `json:"deskripsi_kejadian" binding:"required,min=20"`
	LokasiKejadian    string `json:"lokasi_kejadian"`
	TanggalKejadian   string `json:"tanggal_kejadian"` // format: "2006-01-02"
	IsAnonim          bool   `json:"is_anonim"`
	TingkatUrgensi    string `json:"tingkat_urgensi"   binding:"required"`
}

type UpdateBullyStatusRequest struct {
	Status            string `json:"status"             binding:"required"`
	CatatanPenanganan string `json:"catatan_penanganan"`
}

// Payload yang dikirim ke webhook n8n
type BullyWebhookPayload struct {
	ReportUID       string `json:"report_uid"`
	NPSN            string `json:"npsn"`
	NamaSekolah     string `json:"nama_sekolah"`
	JenisBully      string `json:"jenis_bully"`
	TingkatUrgensi  string `json:"tingkat_urgensi"`
	DeskripsiSingkat string `json:"deskripsi_singkat"` // maks 200 karakter
	IsAnonim        bool   `json:"is_anonim"`
	NamaPelapor     string `json:"nama_pelapor"` // "Anonim" jika is_anonim=true
	LinkDetail      string `json:"link_detail"`
	CreatedAt       string `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Validasi tipe file via magic bytes (bukan ekstensi atau Content-Type header)
// ─────────────────────────────────────────────────────────────────────────────

var allowedMIME = map[string][]byte{
	"image/jpeg": {0xFF, 0xD8, 0xFF},
	"image/png":  {0x89, 0x50, 0x4E, 0x47},
	// WebP: dimulai dengan "RIFF" di byte 0-3 dan "WEBP" di byte 8-11
}

func detectImageMIME(data []byte) (string, bool) {
	if len(data) < 12 {
		return "", false
	}
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg", true
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png", true
	}
	// WebP: "RIFF" + 4 bytes size + "WEBP"
	if len(data) >= 12 &&
		data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P' {
		return "image/webp", true
	}
	return "", false
}

// ─────────────────────────────────────────────────────────────────────────────
// Webhook n8n — dipanggil secara async via goroutine
// ─────────────────────────────────────────────────────────────────────────────

func SendBullyReportWebhook(report models.BullyReport, namaSekolah, namaPelapor string) {
	webhookURL := os.Getenv("N8N_BULLY_WEBHOOK_URL")
	if webhookURL == "" {
		log.Printf("[BullyWebhook] N8N_BULLY_WEBHOOK_URL tidak di-set, skip notifikasi report_uid=%s", report.ReportUID)
		return
	}

	appBaseURL := os.Getenv("APP_BASE_URL")
	if appBaseURL == "" {
		appBaseURL = "https://mentalhealth.web.id"
	}

	// Potong deskripsi jadi maks 200 karakter untuk preview notifikasi
	deskripsiSingkat := report.DeskripsiKejadian
	if len([]rune(deskripsiSingkat)) > 200 {
		runes := []rune(deskripsiSingkat)
		deskripsiSingkat = string(runes[:197]) + "..."
	}

	payload := BullyWebhookPayload{
		ReportUID:        report.ReportUID,
		NPSN:             report.NPSN,
		NamaSekolah:      namaSekolah,
		JenisBully:       report.JenisBully,
		TingkatUrgensi:   report.TingkatUrgensi,
		DeskripsiSingkat: deskripsiSingkat,
		IsAnonim:         report.IsAnonim,
		NamaPelapor:      namaPelapor,
		LinkDetail:       fmt.Sprintf("%s/gurubk/bully-reports/%s", appBaseURL, report.ReportUID),
		CreatedAt:        report.CreatedAt.Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[BullyWebhook] Gagal marshal payload: %v", err)
		return
	}

	webhookSecret := os.Getenv("N8N_WEBHOOK_SECRET")
	const maxRetry = 2
	var lastErr error

	for attempt := 1; attempt <= maxRetry; attempt++ {
		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(body))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if webhookSecret != "" {
			req.Header.Set("X-Webhook-Secret", webhookSecret)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("[BullyWebhook] Percobaan %d gagal: %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Sukses — update flag di DB
			database.DB.Model(&models.BullyReport{}).
				Where("report_uid = ?", report.ReportUID).
				Update("notifikasi_terkirim", true)
			log.Printf("[BullyWebhook] Berhasil kirim notifikasi report_uid=%s", report.ReportUID)
			return
		}
		lastErr = fmt.Errorf("HTTP %d dari n8n", resp.StatusCode)
		log.Printf("[BullyWebhook] Percobaan %d: %v", attempt, lastErr)
		time.Sleep(2 * time.Second)
	}

	log.Printf("[BullyWebhook] Semua percobaan gagal untuk report_uid=%s: %v", report.ReportUID, lastErr)
	// notifikasi_terkirim tetap false — Guru BK akan melihat indikator ⚠️ di dashboard
}

// ─────────────────────────────────────────────────────────────────────────────
// SISWA — Submit laporan bully
// POST /api/siswa/bully-report
// ─────────────────────────────────────────────────────────────────────────────

func SubmitBullyReport(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	// Rate limiting: maks 3 laporan per hari per siswa
	const maxPerDay = 3
	var countToday int64
	today := time.Now().UTC().Truncate(24 * time.Hour)
	database.DB.Model(&models.BullyReport{}).
		Where("pelapor_uid = ? AND created_at >= ?", claims.ID, today).
		Count(&countToday)
	if countToday >= maxPerDay {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": fmt.Sprintf("Anda sudah mengirim %d laporan hari ini. Batas maksimal %d laporan per hari.", countToday, maxPerDay),
		})
		return
	}

	var req SubmitBullyReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	// Validasi jenis_bully
	validJenis := map[string]bool{"Verbal": true, "Fisik": true, "Cyberbullying": true, "Sosial/Pengucilan": true, "Lainnya": true}
	if !validJenis[req.JenisBully] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jenis bully tidak valid"})
		return
	}

	// Validasi tingkat_urgensi
	validUrgensi := map[string]bool{"Rendah": true, "Sedang": true, "Tinggi/Darurat": true}
	if !validUrgensi[req.TingkatUrgensi] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tingkat urgensi tidak valid"})
		return
	}

	// Ambil NPSN siswa
	var student models.Students
	if err := database.DB.Where("students_uid = ?", claims.ID).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data siswa tidak ditemukan"})
		return
	}

	// Parse tanggal kejadian (opsional)
	var tanggalKejadian *time.Time
	if req.TanggalKejadian != "" {
		t, err := time.Parse("2006-01-02", req.TanggalKejadian)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal_kejadian tidak valid (gunakan YYYY-MM-DD)"})
			return
		}
		tanggalKejadian = &t
	}

	report := models.BullyReport{
		PelaporUID:        claims.ID, // Opsi B: selalu disimpan
		NPSN:              student.NPSN,
		JenisBully:        req.JenisBully,
		NamaTerlapor:      req.NamaTerlapor,
		KelasTerlapor:     req.KelasTerlapor,
		DeskripsiKejadian: req.DeskripsiKejadian,
		LokasiKejadian:    req.LokasiKejadian,
		TanggalKejadian:   tanggalKejadian,
		IsAnonim:          req.IsAnonim,
		TingkatUrgensi:    req.TingkatUrgensi,
		Status:            "BARU",
		NotifikasiTerkirim: false,
	}

	if err := database.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan laporan"})
		return
	}

	// Ambil nama sekolah untuk payload webhook
	var namaSekolah string
	database.DB.Table("sekolahs").Where("npsn = ?", student.NPSN).Select("nama_sekolah").Scan(&namaSekolah)

	// Tentukan nama pelapor untuk notifikasi
	namaPelapor := student.NamaLengkap
	if req.IsAnonim {
		namaPelapor = "Anonim"
	}

	// Kirim webhook ke n8n secara ASYNC (tidak memblokir response ke siswa)
	go SendBullyReportWebhook(report, namaSekolah, namaPelapor)

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Laporan berhasil dikirim dan akan segera ditindaklanjuti oleh Guru BK.",
		"report_uid": report.ReportUID,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// SISWA — Daftar laporan milik siswa yang login
// GET /api/siswa/bully-report/my-reports
// ─────────────────────────────────────────────────────────────────────────────

func GetMyBullyReports(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	var reports []models.BullyReport
	if err := database.DB.
		Where("pelapor_uid = ?", claims.ID).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat laporan"})
		return
	}

	// Hapus PelaporUID dari response (tidak perlu ditampilkan ke siswa sendiri)
	type SafeReport struct {
		ReportUID          string     `json:"report_uid"`
		JenisBully         string     `json:"jenis_bully"`
		NamaTerlapor       string     `json:"nama_terlapor"`
		KelasTerlapor      string     `json:"kelas_terlapor"`
		DeskripsiKejadian  string     `json:"deskripsi_kejadian"`
		LokasiKejadian     string     `json:"lokasi_kejadian"`
		TanggalKejadian    *time.Time `json:"tanggal_kejadian"`
		IsAnonim           bool       `json:"is_anonim"`
		TingkatUrgensi     string     `json:"tingkat_urgensi"`
		Status             string     `json:"status"`
		CatatanPenanganan  string     `json:"catatan_penanganan"`
		NotifikasiTerkirim bool       `json:"notifikasi_terkirim"`
		CreatedAt          time.Time  `json:"created_at"`
	}

	var safeReports []SafeReport
	for _, r := range reports {
		safeReports = append(safeReports, SafeReport{
			ReportUID:          r.ReportUID,
			JenisBully:         r.JenisBully,
			NamaTerlapor:       r.NamaTerlapor,
			KelasTerlapor:      r.KelasTerlapor,
			DeskripsiKejadian:  r.DeskripsiKejadian,
			LokasiKejadian:     r.LokasiKejadian,
			TanggalKejadian:    r.TanggalKejadian,
			IsAnonim:           r.IsAnonim,
			TingkatUrgensi:     r.TingkatUrgensi,
			Status:             r.Status,
			CatatanPenanganan:  r.CatatanPenanganan,
			NotifikasiTerkirim: r.NotifikasiTerkirim,
			CreatedAt:          r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"Data": safeReports})
}

// ─────────────────────────────────────────────────────────────────────────────
// SISWA — Upload foto bukti
// POST /api/siswa/bully-report/:report_uid/attachments
// ─────────────────────────────────────────────────────────────────────────────

func UploadBullyAttachment(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	reportUID := c.Param("report_uid")

	// Pastikan laporan milik siswa yang login
	var report models.BullyReport
	if err := database.DB.Where("report_uid = ? AND pelapor_uid = ?", reportUID, claims.ID).First(&report).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Laporan tidak ditemukan atau Anda tidak berhak mengakses laporan ini"})
		return
	}

	// Cek jumlah lampiran yang sudah ada (maks 5)
	var existingCount int64
	database.DB.Model(&models.BullyReportAttachment{}).Where("report_uid = ?", reportUID).Count(&existingCount)
	if existingCount >= 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maksimal 5 foto per laporan sudah tercapai"})
		return
	}

	// Parse multipart form (maks 6 MB per request untuk satu file + overhead)
	if err := c.Request.ParseMultipartForm(6 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal memproses form upload"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan di request"})
		return
	}
	defer file.Close()

	// Batasi ukuran: 5 MB
	const maxSize = 5 * 1024 * 1024
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran file melebihi batas 5 MB"})
		return
	}

	// Baca 12 byte pertama untuk validasi magic bytes
	header12 := make([]byte, 12)
	if _, err := file.Read(header12); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca file"})
		return
	}

	detectedMIME, ok := detectImageMIME(header12)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipe file tidak didukung. Hanya JPEG, PNG, dan WebP yang diizinkan."})
		return
	}

	// Reset pointer dan baca ulang seluruh file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses file"})
		return
	}

	// Tentukan ekstensi berdasarkan MIME yang terdeteksi
	extMap := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}
	ext := extMap[detectedMIME]

	// Buat folder jika belum ada
	uploadDir := filepath.Join("uploads", "bully-evidence", reportUID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat direktori upload"})
		return
	}

	// Nama file unik berdasarkan timestamp
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, fileName)

	// Tulis file ke disk
	outFile, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
		return
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menulis file ke disk"})
		return
	}

	// Simpan metadata ke DB
	attachment := models.BullyReportAttachment{
		ReportUID: reportUID,
		FilePath:  filePath,
		FileType:  detectedMIME,
		FileSize:  header.Size,
	}
	if err := database.DB.Create(&attachment).Error; err != nil {
		// Hapus file jika DB gagal
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan metadata lampiran"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Foto berhasil diupload",
		"attachment_id": attachment.AttachmentID,
		"file_type":     detectedMIME,
		"file_size":     header.Size,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GURU BK — Daftar laporan bully (filter by NPSN guru yang login)
// GET /api/gurubk/bully-reports?status=BARU&urgensi=Tinggi-Darurat&page=1&limit=20
// ─────────────────────────────────────────────────────────────────────────────

func GetGurubkBullyReports(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	// Ambil NPSN guru yang login
	var teacherNPSN string
	if err := database.DB.Table("teachers").
		Where("n_ip = ?", claims.ID).
		Select("npsn").Scan(&teacherNPSN).Error; err != nil || teacherNPSN == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data guru"})
		return
	}

	// Query params filter
	statusFilter  := c.Query("status")
	urgensiFilter := c.Query("urgensi")
	page, _       := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _      := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := database.DB.Model(&models.BullyReport{}).Where("npsn = ?", teacherNPSN)
	if statusFilter != "" {
		query = query.Where("status = ?", strings.ToUpper(statusFilter))
	}
	if urgensiFilter != "" {
		query = query.Where("tingkat_urgensi = ?", urgensiFilter)
	}

	var total int64
	query.Count(&total)

	var reports []models.BullyReport
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat laporan"})
		return
	}

	// Sembunyikan PelaporUID dari response Guru BK (pseudo-anonymous)
	type GurubkReport struct {
		ReportUID          string     `json:"report_uid"`
		JenisBully         string     `json:"jenis_bully"`
		NamaTerlapor       string     `json:"nama_terlapor"`
		KelasTerlapor      string     `json:"kelas_terlapor"`
		DeskripsiKejadian  string     `json:"deskripsi_kejadian"`
		LokasiKejadian     string     `json:"lokasi_kejadian"`
		TanggalKejadian    *time.Time `json:"tanggal_kejadian"`
		IsAnonim           bool       `json:"is_anonim"`
		NamaPelapor        string     `json:"nama_pelapor"` // "Anonim" atau nama asli
		TingkatUrgensi     string     `json:"tingkat_urgensi"`
		Status             string     `json:"status"`
		DitanganiOleh      string     `json:"ditangani_oleh"`
		CatatanPenanganan  string     `json:"catatan_penanganan"`
		NotifikasiTerkirim bool       `json:"notifikasi_terkirim"`
		CreatedAt          time.Time  `json:"created_at"`
	}

	var result []GurubkReport
	for _, r := range reports {
		namaPelapor := "Anonim"
		if !r.IsAnonim && r.PelaporUID != "" {
			// Ambil nama siswa pelapor
			var s models.Students
			if err := database.DB.Where("students_uid = ?", r.PelaporUID).Select("nama_lengkap").First(&s).Error; err == nil {
				namaPelapor = s.NamaLengkap
			}
		}
		result = append(result, GurubkReport{
			ReportUID:          r.ReportUID,
			JenisBully:         r.JenisBully,
			NamaTerlapor:       r.NamaTerlapor,
			KelasTerlapor:      r.KelasTerlapor,
			DeskripsiKejadian:  r.DeskripsiKejadian,
			LokasiKejadian:     r.LokasiKejadian,
			TanggalKejadian:    r.TanggalKejadian,
			IsAnonim:           r.IsAnonim,
			NamaPelapor:        namaPelapor,
			TingkatUrgensi:     r.TingkatUrgensi,
			Status:             r.Status,
			DitanganiOleh:      r.DitanganiOleh,
			CatatanPenanganan:  r.CatatanPenanganan,
			NotifikasiTerkirim: r.NotifikasiTerkirim,
			CreatedAt:          r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"Data":  result,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GURU BK — Detail satu laporan
// GET /api/gurubk/bully-reports/:report_uid
// ─────────────────────────────────────────────────────────────────────────────

func GetGurubkBullyReportDetail(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	reportUID := c.Param("report_uid")

	// Verifikasi NPSN guru cocok dengan NPSN laporan
	var teacherNPSN string
	database.DB.Table("teachers").Where("n_ip = ?", claims.ID).Select("npsn").Scan(&teacherNPSN)

	var report models.BullyReport
	if err := database.DB.Where("report_uid = ? AND npsn = ?", reportUID, teacherNPSN).First(&report).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laporan tidak ditemukan"})
		return
	}

	// Ambil lampiran
	var attachments []models.BullyReportAttachment
	database.DB.Where("report_uid = ?", reportUID).Find(&attachments)

	// Bangun response — sembunyikan PelaporUID untuk Guru BK
	namaPelapor := "Anonim"
	if !report.IsAnonim && report.PelaporUID != "" {
		var s models.Students
		if err := database.DB.Where("students_uid = ?", report.PelaporUID).Select("nama_lengkap, nisn, kelas").First(&s).Error; err == nil {
			namaPelapor = fmt.Sprintf("%s (NISN: %s, Kelas: %s)", s.NamaLengkap, s.NISN, s.Kelas)
		}
	}

	// Attachment tanpa file_path (keamanan: path server tidak diekspos ke client)
	type SafeAttachment struct {
		AttachmentID int64     `json:"attachment_id"`
		FileType     string    `json:"file_type"`
		FileSize     int64     `json:"file_size"`
		UploadedAt   time.Time `json:"uploaded_at"`
	}
	var safeAttachments []SafeAttachment
	for _, a := range attachments {
		safeAttachments = append(safeAttachments, SafeAttachment{
			AttachmentID: a.AttachmentID,
			FileType:     a.FileType,
			FileSize:     a.FileSize,
			UploadedAt:   a.UploadedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"Data": gin.H{
			"report_uid":          report.ReportUID,
			"jenis_bully":         report.JenisBully,
			"nama_terlapor":       report.NamaTerlapor,
			"kelas_terlapor":      report.KelasTerlapor,
			"deskripsi_kejadian":  report.DeskripsiKejadian,
			"lokasi_kejadian":     report.LokasiKejadian,
			"tanggal_kejadian":    report.TanggalKejadian,
			"is_anonim":           report.IsAnonim,
			"nama_pelapor":        namaPelapor,
			"tingkat_urgensi":     report.TingkatUrgensi,
			"status":              report.Status,
			"ditangani_oleh":      report.DitanganiOleh,
			"catatan_penanganan":  report.CatatanPenanganan,
			"notifikasi_terkirim": report.NotifikasiTerkirim,
			"created_at":          report.CreatedAt,
			"attachments":         safeAttachments,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GURU BK — Update status laporan
// PATCH /api/gurubk/bully-reports/:report_uid/status
// ─────────────────────────────────────────────────────────────────────────────

func UpdateBullyReportStatus(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	reportUID := c.Param("report_uid")

	var req UpdateBullyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	validStatus := map[string]bool{"BARU": true, "DITINDAKLANJUTI": true, "SELESAI": true, "DITOLAK": true}
	if !validStatus[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status tidak valid"})
		return
	}

	// Verifikasi NPSN guru cocok
	var teacherNPSN string
	database.DB.Table("teachers").Where("n_ip = ?", claims.ID).Select("npsn").Scan(&teacherNPSN)

	var report models.BullyReport
	if err := database.DB.Where("report_uid = ? AND npsn = ?", reportUID, teacherNPSN).First(&report).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laporan tidak ditemukan"})
		return
	}

	updates := map[string]interface{}{
		"status":             req.Status,
		"catatan_penanganan": req.CatatanPenanganan,
		"ditangani_oleh":     claims.ID,
		"updated_at":         time.Now(),
	}

	if err := database.DB.Model(&report).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update status laporan"})
		return
	}

	// Kirim notifikasi in-app ke siswa pelapor jika TIDAK anonim
	if !report.IsAnonim && report.PelaporUID != "" && req.Status != report.Status {
		statusLabel := map[string]string{
			"BARU":            "Baru",
			"DITINDAKLANJUTI": "Sedang Ditindaklanjuti",
			"SELESAI":         "Selesai",
			"DITOLAK":         "Ditolak",
		}
		label := statusLabel[req.Status]
		go createStudentNotification(
			report.PelaporUID,
			"BULLY_STATUS_CHANGE",
			"Status Laporan Diperbarui",
			fmt.Sprintf("Laporan bully yang kamu kirim kini berstatus: %s.", label),
			report.ReportUID,
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status laporan berhasil diperbarui"})
}


// ─────────────────────────────────────────────────────────────────────────────
// GURU BK — Akses foto bukti (via attachment_id)
// GET /api/gurubk/bully-reports/:report_uid/attachments/:id
// ─────────────────────────────────────────────────────────────────────────────

func GetBullyAttachmentGurubk(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	reportUID    := c.Param("report_uid")
	attachmentID := c.Param("id")

	// Verifikasi NPSN guru cocok dengan laporan
	var teacherNPSN string
	database.DB.Table("teachers").Where("n_ip = ?", claims.ID).Select("npsn").Scan(&teacherNPSN)

	var report models.BullyReport
	if err := database.DB.Where("report_uid = ? AND npsn = ?", reportUID, teacherNPSN).First(&report).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak"})
		return
	}

	// Ambil metadata lampiran
	var attachment models.BullyReportAttachment
	if err := database.DB.Where("attachment_id = ? AND report_uid = ?", attachmentID, reportUID).First(&attachment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lampiran tidak ditemukan"})
		return
	}

	// Validasi file ada di disk
	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File tidak ditemukan di server"})
		return
	}

	c.Header("Content-Type", attachment.FileType)
	c.Header("Cache-Control", "private, max-age=3600")
	c.File(attachment.FilePath)
}

// ─────────────────────────────────────────────────────────────────────────────
// SISWA — Akses foto bukti miliknya sendiri
// GET /api/siswa/bully-report/:report_uid/attachments/:id
// ─────────────────────────────────────────────────────────────────────────────

func GetBullyAttachmentSiswa(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
		return
	}

	reportUID    := c.Param("report_uid")
	attachmentID := c.Param("id")

	// Pastikan laporan milik siswa yang login
	var report models.BullyReport
	if err := database.DB.Where("report_uid = ? AND pelapor_uid = ?", reportUID, claims.ID).First(&report).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak"})
		return
	}

	var attachment models.BullyReportAttachment
	if err := database.DB.Where("attachment_id = ? AND report_uid = ?", attachmentID, reportUID).First(&attachment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lampiran tidak ditemukan"})
		return
	}

	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File tidak ditemukan di server"})
		return
	}

	c.Header("Content-Type", attachment.FileType)
	c.Header("Cache-Control", "private, max-age=3600")
	c.File(attachment.FilePath)
}
