package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// backupScriptPath mengembalikan path skrip bash dari env var atau default.
func backupScriptPath() string {
	if p := os.Getenv("BACKUP_SCRIPT_PATH"); p != "" {
		return p
	}
	return "./scripts/backup/backup_sindas.sh"
}

// ─── TriggerBackup ────────────────────────────────────────────────────────────
// POST /api/admin/backup/trigger
// Body: { "type": "weekly" | "annual" }
//
// Rate limit:
//   1. Global: tolak jika ada PENDING/RUNNING dalam 1 jam terakhir (server load)
//   2. Per-Admin: tolak jika Admin ini sudah trigger ≥ 3 kali hari ini

type TriggerBackupRequest struct {
	Type string `json:"type" binding:"required"`
}

func TriggerBackup(c *gin.Context) {
	// Pastikan user adalah Admin
	userType, _ := c.Get("user_type")
	userTypeStr, ok := userType.(string)
	if !ok || (userTypeStr != "admin" && userTypeStr != "Administrator" && userTypeStr != "administrator") {
		c.JSON(http.StatusForbidden, gin.H{
			"Status":  "Error",
			"Message": "Hanya Admin yang dapat memicu backup manual",
		})
		return
	}

	adminUID, _ := c.Get("user_uid")
	adminUIDStr := fmt.Sprintf("%v", adminUID)

	var req TriggerBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Type != "weekly" && req.Type != "annual") {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Tipe backup tidak valid. Gunakan 'weekly' atau 'annual'",
		})
		return
	}

	// ── Auto-cleanup: Reset PENDING/RUNNING yang terbengkalai > 30 menit ──────
	// Ini mencegah rate-limit "racun" akibat proses yang mati tanpa cleanup DB
	staleCutoff := time.Now().Add(-30 * time.Minute)
	staleMsg := "Otomatis dibatalkan: timeout 30 menit"
	finishTime := time.Now()
	database.DB.Model(&models.BackupJob{}).
		Where("status IN ('PENDING','RUNNING') AND created_at < ?", staleCutoff).
		Updates(map[string]interface{}{
			"status":        "FAILED",
			"error_message": staleMsg,
			"finished_at":   finishTime,
		})

	// ── Rate limit 1: ada PENDING/RUNNING dalam 15 menit terakhir? ────────────
	var runningCount int64
	fifteenMinAgo := time.Now().Add(-15 * time.Minute)
	database.DB.Model(&models.BackupJob{}).
		Where("status IN ('PENDING','RUNNING') AND created_at >= ?", fifteenMinAgo).
		Count(&runningCount)

	if runningCount > 0 {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"Status":  "Error",
			"Message": "Backup sedang berjalan atau baru saja dipicu, coba lagi nanti (cooldown 15 menit)",
		})
		return
	}

	// ── Rate limit 2: Admin ini sudah ≥ 3 kali hari ini? ─────────────────────
	todayStart := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	database.DB.Model(&models.BackupJob{}).
		Where("triggered_by = ? AND created_at >= ?", adminUIDStr, todayStart).
		Count(&todayCount)

	if todayCount >= 3 {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"Status":  "Error",
			"Message": "Batas backup manual harian tercapai (3x/hari). Coba lagi besok",
		})
		return
	}

	// ── Insert record baru ke backup_jobs ─────────────────────────────────────
	jobUID := uuid.New().String()
	triggeredBy := adminUIDStr
	job := models.BackupJob{
		JobUID:      jobUID,
		Type:        req.Type,
		Status:      "PENDING",
		TriggeredBy: &triggeredBy,
		CreatedAt:   time.Now(),
	}
	if err := database.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal membuat record backup job",
			"Error":   err.Error(),
		})
		return
	}

	// ── Helper: tandai job sebagai FAILED di DB ────────────────────────────────
	markFailed := func(jobUID, reason string) {
		now := time.Now()
		database.DB.Model(&models.BackupJob{}).
			Where("job_uid = ?", jobUID).
			Updates(map[string]interface{}{
				"status":        "FAILED",
				"error_message": reason,
				"finished_at":   now,
			})
	}

	// ── Jalankan skrip backup di background (non-blocking) ────────────────────
	if runtime.GOOS == "windows" {
		// Mock Mode untuk testing di Windows lokal tanpa bash/Linux toolchain
		go func(jobUID string, backupType string) {
			now := time.Now()
			// Update status ke RUNNING dulu
			database.DB.Model(&models.BackupJob{}).
				Where("job_uid = ?", jobUID).
				Updates(map[string]interface{}{
					"status":     "RUNNING",
					"started_at": now,
				})

			// Simulasikan pengerjaan backup selama 3 detik
			time.Sleep(3 * time.Second)

			// Buat dummy backup file di folder scratch lokal
			dummyDir := "./scratch"
			_ = os.MkdirAll(dummyDir, 0755)
			dummyPath := dummyDir + "/dummy_backup.dump.gpg"
			_ = os.WriteFile(dummyPath, []byte("DUMMY ENCRYPTED BACKUP FILE FOR TESTING ON WINDOWS"), 0644)

			finishTime := time.Now()
			database.DB.Model(&models.BackupJob{}).
				Where("job_uid = ?", jobUID).
				Updates(map[string]interface{}{
					"status":      "SUCCESS",
					"file_path":   dummyPath,
					"file_size":   "12 KB",
					"finished_at": finishTime,
				})
		}(jobUID, req.Type)

		c.JSON(http.StatusAccepted, gin.H{
			"Status":  "Accepted",
			"Message": "Backup manual dipicu (Windows Mock Mode). Pantau status via job_uid",
			"Data": gin.H{
				"job_uid": jobUID,
				"type":    req.Type,
				"status":  "PENDING",
			},
		})
		return
	}

	// ── Linux/VPS: jalankan bash script di goroutine bertracking ──────────────
	scriptPath := backupScriptPath()
	go func(jobUID, scriptPath, backupType string) {
		// Update ke RUNNING
		startTime := time.Now()
		database.DB.Model(&models.BackupJob{}).
			Where("job_uid = ?", jobUID).
			Updates(map[string]interface{}{
				"status":     "RUNNING",
				"started_at": startTime,
			})

		jobIDArg := "--job-id=" + jobUID
		cmd := exec.Command("bash", scriptPath, backupType, jobIDArg)
		cmd.Stdout = nil
		cmd.Stderr = nil

		// Timeout 30 menit
		done := make(chan error, 1)
		if err := cmd.Start(); err != nil {
			markFailed(jobUID, "Gagal memulai proses backup: "+err.Error())
			return
		}

		go func() { done <- cmd.Wait() }()

		timeout := time.After(30 * time.Minute)
		select {
		case err := <-done:
			if err != nil {
				markFailed(jobUID, "Proses backup keluar dengan error: "+err.Error())
			}
			// Jika sukses, skrip bash seharusnya sudah update DB sendiri.
			// Namun jika status masih RUNNING (skrip tidak update), tandai FAILED.
			var current models.BackupJob
			if dbErr := database.DB.Where("job_uid = ?", jobUID).First(&current).Error; dbErr == nil {
				if current.Status == "RUNNING" {
					markFailed(jobUID, "Skrip backup selesai tapi tidak memperbarui status database")
				}
			}
		case <-timeout:
			// Kill proses yang overtime
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			markFailed(jobUID, "Timeout: proses backup melebihi batas waktu 30 menit")
		}
	}(jobUID, scriptPath, req.Type)

	c.JSON(http.StatusAccepted, gin.H{
		"Status":  "Accepted",
		"Message": "Backup manual berhasil dipicu. Pantau status via job_uid",
		"Data": gin.H{
			"job_uid": jobUID,
			"type":    req.Type,
			"status":  "PENDING",
		},
	})
}

// ─── GetBackupJobStatus ────────────────────────────────────────────────────────
// GET /api/admin/backup/jobs/:job_uid

type BackupJobPublic struct {
	JobUID            string     `json:"job_uid"`
	Type              string     `json:"type"`
	Status            string     `json:"status"`
	TriggeredBy       *string    `json:"triggered_by"`
	TriggeredByName   string     `json:"triggered_by_name"`
	FileSize          *string    `json:"file_size"`
	HasFile           bool       `json:"has_file"`   // true jika ada file untuk didownload
	StartedAt         *time.Time `json:"started_at"`
	FinishedAt        *time.Time `json:"finished_at"`
	ErrorMessage      *string    `json:"error_message"`
	CreatedAt         time.Time  `json:"created_at"`
	DurationSeconds   *int64     `json:"duration_seconds"`
}

func buildPublicJob(job models.BackupJob) BackupJobPublic {
	pub := BackupJobPublic{
		JobUID:       job.JobUID,
		Type:         job.Type,
		Status:       job.Status,
		TriggeredBy:  job.TriggeredBy,
		FileSize:     job.FileSize,
		HasFile:      job.FilePath != nil && *job.FilePath != "",
		StartedAt:    job.StartedAt,
		FinishedAt:   job.FinishedAt,
		ErrorMessage: job.ErrorMessage,
		CreatedAt:    job.CreatedAt,
	}

	// Hitung durasi jika ada start + finish
	if job.StartedAt != nil && job.FinishedAt != nil {
		dur := int64(job.FinishedAt.Sub(*job.StartedAt).Seconds())
		pub.DurationSeconds = &dur
	}

	// Nama Admin yang trigger (jika bukan cron)
	if job.TriggeredBy != nil {
		var admin models.Administrator
		if err := database.DB.Where("admin_uid = ?", *job.TriggeredBy).First(&admin).Error; err == nil {
			pub.TriggeredByName = admin.NamaLengkap
		}
	}

	return pub
}

func GetBackupJobStatus(c *gin.Context) {
	jobUID := c.Param("job_uid")

	var job models.BackupJob
	if err := database.DB.Where("job_uid = ?", jobUID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Job backup tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data":   buildPublicJob(job),
	})
}

// ─── ListBackupJobs ───────────────────────────────────────────────────────────
// GET /api/admin/backup/jobs?page=1&limit=20

func ListBackupJobs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	var total int64
	database.DB.Model(&models.BackupJob{}).Count(&total)

	var jobs []models.BackupJob
	database.DB.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs)

	result := make([]BackupJobPublic, 0, len(jobs))
	for _, j := range jobs {
		result = append(result, buildPublicJob(j))
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data": gin.H{
			"jobs":        result,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// ─── DownloadBackupFile ───────────────────────────────────────────────────────
// GET /api/admin/backup/download/:job_uid
//
// Menyajikan file .dump.gpg terenkripsi untuk diunduh Admin.
// file_path tidak pernah di-expose ke client — hanya job_uid yang digunakan.

func DownloadBackupFile(c *gin.Context) {
	jobUID := c.Param("job_uid")

	var job models.BackupJob
	if err := database.DB.Where("job_uid = ? AND status = 'SUCCESS'", jobUID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "File backup tidak ditemukan atau proses belum selesai",
		})
		return
	}

	if job.FilePath == nil || *job.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Path file backup tidak tersedia",
		})
		return
	}

	filePath := *job.FilePath
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusGone, gin.H{
			"Status":  "Error",
			"Message": "File backup sudah tidak ada di server (mungkin sudah melewati masa retensi)",
		})
		return
	}

	// Ambil nama file saja untuk Content-Disposition
	fileName := fmt.Sprintf("backup_sindas_%s_%s.dump.gpg", job.Type, job.CreatedAt.Format("20060102"))

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("X-Content-Type-Options", "nosniff")
	c.File(filePath)
}