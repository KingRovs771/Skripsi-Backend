package models

type Pertanyaan struct {
	PertanyaanId int    `gorm:"primaryKey;uniqueIndex" json:"pertanyaan_id"`
	JenisKuisId  string `gorm:"type:varchar" json:"jenis_kuis_id"`
	Description  string `grom:"type:text" json:"description"`
}
