package models

import "time"

type Gejala struct {
	GejalaID    int64     `gorm:"primaryKey;uniqueIndex" json:"gejala_id"`
	GejalaUID   string    `gorm:"type:varchar" json:"gejala_uid"`
	KodeGejala  string    `gorm:"type:varchar(30)" json:"kode_gejala"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt    time.Time `gorm:"type:timestamp" json:"update_at"`
}
