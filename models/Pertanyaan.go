package models

import "time"

type Pertanyaan struct {
	PertanyaanId       int       `gorm:"primaryKey;uniqueIndex" json:"pertanyaan_id"`
	PertanyaanUID      string    `gorm:"varchar(255)" json:"pertanyaan_uid"`
	KodePertanyaan     string    `gorm:"varchar(255)" json:"kode_pertanyaan"`
	KategoriPertanyaan string    `gorm:"type:varchar" json:"kategori_pertanyaan"`
	Pertanyaan         string    `gorm:"text" json:"pertanyaan"`
	Bobot              float64   `gorm:"type:decimal(10,2)" json:"bobot"`
	CreatedAt          time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt           time.Time `gorm:"type:timestamp" json:"update_at"`
}
