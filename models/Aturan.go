package models

import "time"

type Aturan struct {
	AturanID   int64     `gorm:"primaryKey;uniqueIndex" json:"aturan_id"`
	AturanUID  string    `gorm:"type:varchar(255)" json:"aturan_uid"`
	KodeAturan string    `gorm:"type:varchar(30)" json:"kode_aturan"`
	Rules      string    `gorm:"type:int" json:"rules"`
	CreatedAt  time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt   time.Time `gorm:"type:timestamp" json:"update_at"`
}
