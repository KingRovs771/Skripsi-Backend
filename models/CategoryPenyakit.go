package models

import (
	"Skripsi-Backend/database"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryPenyakit struct {
	CategoryPenyakitId  int64     `gorm:"primaryKey;uniqueIndex" json:"category_penyakit_id"`
	CategoryPenyakitUID string    `gorm:"type:varchar(255)" json:"category_penyakit_uid"`
	KodeCategory        string    `gorm:"type:varchar(40)" json:"kode_category"`
	NamaCategory        string    `gorm:"type:varchar(20)" json:"nama_category"`
	Deskripsi           string    `gorm:"type:text" json:"deskripsi"`
	CreatedAt           time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt            time.Time `gorm:"type:timestamp" json:"update_at"`
}

func (c *CategoryPenyakit) BeforeCreate(db *gorm.DB) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	c.CategoryPenyakitUID = uid.String()

	c.NamaCategory = html.EscapeString(strings.TrimSpace(c.NamaCategory))
	c.Deskripsi = html.EscapeString(strings.TrimSpace(c.Deskripsi))
	c.CreatedAt = time.Now()
	return nil
}

func (c *CategoryPenyakit) BeforeUpdate(db *gorm.DB) error {
	c.NamaCategory = html.EscapeString(strings.TrimSpace(c.NamaCategory))
	c.Deskripsi = html.EscapeString(strings.TrimSpace(c.Deskripsi))
	c.UpdateAt = time.Now()
	return nil
}

func (c *CategoryPenyakit) SaveCategoryPenyakits() (CategoryPenyakit, error) {
	var categorypenyakit CategoryPenyakit

	err := database.DB.Create(&categorypenyakit).Error
	if err != nil {
		return categorypenyakit, err
	}
	return categorypenyakit, nil
}

func GetAllCategoryPenyakits() ([]CategoryPenyakit, error) {
	var categorypenyakits []CategoryPenyakit
	err := database.DB.Find(&categorypenyakits).Error
	if err != nil {
		return categorypenyakits, err
	}
	return categorypenyakits, nil
}

func GetCategoryPenyakitByUID(uid string) (CategoryPenyakit, error) {
	var categorypenyakit CategoryPenyakit
	err := database.DB.Where("category_penyakit_uid = ?", uid).First(&categorypenyakit).Error
	if err != nil {
		return categorypenyakit, err
	}
	return categorypenyakit, nil
}

func (c *CategoryPenyakit) UpdateCategoryPenyakits(uid string) error {
	err := database.DB.Where("category_penyakit_uid = ?", uid).Updates(c).Error
	return err
}

func DeleteCategoryPenyakits(uid string) error {
	err := database.DB.Where("category_penyakit_uid =?", uid).Delete(&CategoryPenyakit{}).Error
	return err
}
