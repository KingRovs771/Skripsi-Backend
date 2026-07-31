package models

import (
	"Skripsi-Backend/crypto"
	"time"

	"gorm.io/gorm"
)

type StudentFeedback struct {
	FeedbackId     int64     `gorm:"primaryKey;autoIncrement" json:"feedback_id"`
	TestSessionUID string    `gorm:"type:varchar(255);index" json:"test_session_uid"`
	CeritaSiswa    string    `gorm:"type:text" json:"cerita_siswa"`
	CreatedAt      time.Time `gorm:"type:timestamp" json:"created_at"`
}

// ─── GORM Hooks untuk Enkripsi At-Rest (UU PDP) ──────────────────────────────

func (sf *StudentFeedback) BeforeSave(tx *gorm.DB) error {
	if sf.CeritaSiswa != "" && sf.CeritaSiswa != "[Dianonimkan sesuai kebijakan retensi data]" {
		encrypted, err := crypto.EncryptField(sf.CeritaSiswa)
		if err != nil {
			return err
		}
		sf.CeritaSiswa = encrypted
	}
	return nil
}

func (sf *StudentFeedback) AfterFind(tx *gorm.DB) error {
	if sf.CeritaSiswa != "" {
		decrypted, err := crypto.DecryptField(sf.CeritaSiswa)
		if err == nil {
			sf.CeritaSiswa = decrypted
		}
	}
	return nil
}

func (sf *StudentFeedback) AfterSave(tx *gorm.DB) error {
	return sf.AfterFind(tx)
}
