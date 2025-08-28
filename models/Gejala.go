package models

type Gejala struct {
	IdGejala    int64  `gorm:"primaryKey;uniqueIndex" json:"id_gejala"`
	KodeGejala  string `gorm:"type:varchar" json:"kode_gejala"`
	Description string `gorm:"type:text" json:"description"`
}