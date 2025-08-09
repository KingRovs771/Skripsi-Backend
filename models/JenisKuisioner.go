package models

type JenisKuisioner struct {
	JenisKuisionerId   int64  `gorm:"primaryKey" json:"jenis_kuisioner_id"`
	NamaJenisKuisioner string `gorm:"type:varchar(50)" json:"nama_jenis_kuisioner"`
	Deskripsi          string `gorm:"type:text" json:"deskripsi"`
}
