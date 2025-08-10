package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	CategoryId   int64  `gorm:"primaryKey;uniqueIndex" json:"category_id"`
	CategoryUID  string `gorm:"type:varchar(255)" json:"category_uid"`
	NameCategory string `gorm:"type:varchar(100)" json:"name_category"`
	Description  string `gorm:"type:text" json:"description"`
}

func (category *Category) BeforeCreate(tx *gorm.DB) (err error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	category.CategoryUID = newUUID.String()
	return
}
