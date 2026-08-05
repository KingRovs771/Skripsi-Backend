package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StudentNotification menyimpan notifikasi in-app untuk siswa.
// Digunakan saat ini untuk pemberitahuan perubahan status laporan bully.
type StudentNotification struct {
	NotifID    int64     `gorm:"primaryKey;autoIncrement"           json:"notif_id"`
	NotifUID   string    `gorm:"type:varchar(255);uniqueIndex"      json:"notif_uid"`
	StudentUID string    `gorm:"type:varchar(255);not null;index"   json:"student_uid"` // FK ke students.students_uid
	Type       string    `gorm:"type:varchar(50);not null"          json:"type"`         // BULLY_STATUS_CHANGE, dll.
	Title      string    `gorm:"type:varchar(200);not null"         json:"title"`
	Message    string    `gorm:"type:text;not null"                 json:"message"`
	RelatedUID string    `gorm:"type:varchar(255)"                  json:"related_uid"`  // report_uid terkait
	IsRead     bool      `gorm:"not null;default:false"             json:"is_read"`
	CreatedAt  time.Time `gorm:"autoCreateTime"                     json:"created_at"`
}

func (n *StudentNotification) BeforeCreate(tx *gorm.DB) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	n.NotifUID = uid.String()
	return nil
}
