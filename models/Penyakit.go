package models

import "time"

type Penyakit struct {
	PenyakitID      int64     `gorm:"primaryKey;uniqueIndex" json:"penyakit_id"`
	PenyakitUID     string    `gorm:"type:varchar" json:"penyakit_uid"`
	KodePenyakit    string    `gorm:"type:varchar(30)" json:"kode_penyakit"`
	NamaPenyakit    string    `gorm:"type:varchar" json:"nama_penyakit"`
	Description     string    `gorm:"type:text" json:"description"`
	SaranPenanganan string    `gorm:"type:text" json:"saran_penanganan"`
	CreatedAt       time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt        time.Time `gorm:"type:timestamp" json:"update_at"`
}
