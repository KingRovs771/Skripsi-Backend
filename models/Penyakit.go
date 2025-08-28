package models

type Penyakit struct {
	IdPenyakit      int64  `gorm:"primaryKey;uniqueIndex" json:"id_penyakit"`
	KodePenyakit    string `gorm:"type:varchar" json:"kode_penyakit"`
	NamaPenyakit    string `gorm:"type:varchar" json:"nama_penyakit"`
	Description     string `gorm:"type:text" json:"description"`
	SaranPenanganan string `gorm:"type:text" json:"saran_penanganan"`
}