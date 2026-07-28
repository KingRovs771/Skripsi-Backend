package models

import "time"

// BackupJob menyimpan riwayat semua proses backup — baik yang dipicu
// otomatis oleh cron (TriggeredBy = nil) maupun manual oleh Admin.
type BackupJob struct {
	JobId        int64      `gorm:"primaryKey;autoIncrement"         json:"job_id"`
	JobUID       string     `gorm:"type:varchar(255);uniqueIndex"    json:"job_uid"`
	Type         string     `gorm:"type:varchar(20)"                 json:"type"`          // "weekly" | "annual"
	Status       string     `gorm:"type:varchar(20);default:'PENDING'" json:"status"`      // PENDING | RUNNING | SUCCESS | FAILED
	TriggeredBy  *string    `gorm:"type:varchar(255)"                json:"triggered_by"`  // NULL = cron; diisi = admin_uid
	FilePath     *string    `gorm:"type:text"                        json:"-"`             // Tidak di-expose ke frontend
	FileSize     *string    `gorm:"type:varchar(30)"                 json:"file_size"`
	StartedAt    *time.Time `gorm:"type:timestamp"                   json:"started_at"`
	FinishedAt   *time.Time `gorm:"type:timestamp"                   json:"finished_at"`
	ErrorMessage *string    `gorm:"type:text"                        json:"error_message"`
	CreatedAt    time.Time  `gorm:"type:timestamp"                   json:"created_at"`
}