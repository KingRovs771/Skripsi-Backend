package models

type Kuisioner struct {
	KuisionerId   int64  `gorm:"primaryKey" json:"kuisioner_id"`
	JenisKuisId   int64  `gorm:"int" json:"jenis_kuis_id"`
	KodeKuisioner string `gorm:"type:varchar(20)" json:"kode_kuisioner"`
	KuisionerText string `gorm:"type:text" json:"kuisioner_text"`
}
