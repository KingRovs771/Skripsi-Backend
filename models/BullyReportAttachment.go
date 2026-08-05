package models

import "time"

// BullyReportAttachment menyimpan metadata foto bukti laporan bully.
// File fisiknya disimpan di folder uploads/bully-evidence/ di VPS backend.
// Akses file melalui signed endpoint (bukan URL publik langsung).
type BullyReportAttachment struct {
	AttachmentID int64     `gorm:"primaryKey;autoIncrement" json:"attachment_id"`
	ReportUID    string    `gorm:"type:varchar(255);not null;index" json:"report_uid"` // FK ke bully_reports.report_uid
	FilePath     string    `gorm:"type:varchar(500);not null" json:"file_path"`        // relative path di server
	FileType     string    `gorm:"type:varchar(50);not null" json:"file_type"`         // MIME: image/jpeg, image/png, image/webp
	FileSize     int64     `gorm:"not null" json:"file_size"`                          // bytes
	UploadedAt   time.Time `gorm:"autoCreateTime" json:"uploaded_at"`
}
