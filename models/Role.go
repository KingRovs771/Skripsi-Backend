package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	RoleId      int64     `gorm:"primaryKey;uniqueIndex" json:"role_id"`
	RoleUID     string    `gorm:"type:varchar(255)" json:"role_uid"`
	RoleName    string    `gorm:"type:varchar(40)" json:"role_name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp" json:"updated_at"`
}

func (r *Role) BeforeSave(tx *gorm.DB) (err error) {
	if r.RoleUID == "" {
		r.RoleUID = uuid.New().String()
	}
	return
}
