package models

type DetailAturan struct {
	IdDetailAturan int64 `gorm:"primaryKey;uniqueIndex" json:"id_detail_aturan"`
	IdAturan       int64 `gorm:"type:int" json:"id_aturan"`
	IdGejala       int64 `gorm:"type:int" json:"id_gejala"`
}