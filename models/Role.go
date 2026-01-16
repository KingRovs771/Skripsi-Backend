package models

import "time"

type Role struct {
	RoleId      int64     `gorm:"primaryKey;uniqueIndex" json:"role_id"`
	RoleUID     string    `gorm:"type:varchar(255)" json:"role_uid"`
	RoleName    string    `gorm:"type:varchar(40)" json:"role_name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt    time.Time `gorm:"type:timestamp" json:"update_at"`
}
