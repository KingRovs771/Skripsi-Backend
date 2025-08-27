package models

type Kuisioner struct {
	JenisKuisionerId   int64  `gorm:"primaryKey;uniqueIndex" json:"jenis_kuisioner_id"`
	NamaJenisKuisioner string `gorm:"type:varchar(20)" json:"nama_jenis_kuisioner"`
	Deskripsi          string `gorm:"type:text" json:"deskripsi"`
}
