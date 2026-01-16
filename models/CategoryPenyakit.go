package models

import "time"

type CategoryPenyakit struct {
	CategoryPenyakitId  int64     `gorm:"primaryKey;uniqueIndex" json:"category_penyakit_id"`
	CategoryPenyakitUID string    `gorm:"type:varchar(255)" json:"category_penyakit_uid"`
	NamaCategory        string    `gorm:"type:varchar(20)" json:"nama_category"`
	Deskripsi           string    `gorm:"type:text" json:"deskripsi"`
	CreatedAt           time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt            time.Time `gorm:"type:timestamp" json:"update_at"`
}
