package models

import "time"

type Aturan struct {
	AturanID       int64     `gorm:"primaryKey;uniqueIndex" json:"aturan_id"`
	AturanUID      string    `gorm:"type:varchar(255)" json:"aturan_uid"`
	KodePenyakit   string    `gorm:"type:varchar(255)" json:"kode_penyakit"`
	KodePertanyaan string    `gorm:"type:varchar(255)" json:"kode_pertanyaan"`
	MinValue       int64     `gorm:"type:bigint" json:"min_value"`
	IsMandatory    int64     `gorm:"type:bigint" json:"is_mandatory"`
	CreatedAt      time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt       time.Time `gorm:"type:timestamp" json:"update_at"`
}
