package models

type Role struct {
	RoleId      int64  `gorm:"primaryKey;uniqueIndex" json:"role_id"`
	RoleName    string `gorm:"type:varchar(40)" json:"role_name"`
	Description string `gorm:"type:text" json:"description"`
}
