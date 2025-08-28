package models

type Aturan struct {
	IdAturan             int64  `gorm:"primaryKey;uniqueIndex" json:"id_aturan"`
	IdPenyakitKesimpulan string `gorm:"type:int" json:"id_penyakit_kesimpulan"`
}
