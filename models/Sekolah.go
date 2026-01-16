package models

import "time"

type Sekolah struct {
	SekolahId     int64     `gorm:"primaryKey;uniqueIndex" json:"sekolah_id"`
	SekolahUID    string    `gorm:"type:varchar(255)" json:"sekolah_uid"`
	NPSN          int64     `gorm:"type:int" json:"npsn"`
	NamaSekolah   string    `gorm:"type:varchar(100)" json:"nama_sekolah"`
	Jenjang       string    `gorm:"type:varchar(20)" json:"jenjang"`
	AlamatSekolah string    `gorm:"type:text" json:"alamat_sekolah"`
	CreatedAt     time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt      time.Time `gorm:"type:timestamp" json:"update_at"`
}
