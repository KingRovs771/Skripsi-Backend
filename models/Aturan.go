package models

import (
	"Skripsi-Backend/database"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

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

func (a *Aturan) BeforeCreate(db *gorm.DB) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	a.AturanUID = uid.String()

	a.KodePenyakit = html.EscapeString(strings.TrimSpace(a.KodePenyakit))
	a.KodePertanyaan = html.EscapeString(strings.TrimSpace(a.KodePertanyaan))
	a.CreatedAt = time.Now()

	return nil
}

func (a *Aturan) BeforeUpdate(db *gorm.DB) error {
	a.KodePenyakit = html.EscapeString(strings.TrimSpace(a.KodePenyakit))
	a.KodePertanyaan = html.EscapeString(strings.TrimSpace(a.KodePertanyaan))
	a.UpdateAt = time.Now()
	return nil
}

func (a *Aturan) SaveAturan() (Aturan, error) {
	var aturan Aturan
	err := database.DB.Create(&aturan).Error
	if err != nil {
		return aturan, err
	}
	return aturan, nil
}

func GetAllAturan() ([]Aturan, error) {
	var aturans []Aturan
	err := database.DB.Find(&aturans).Error
	if err != nil {
		return aturans, err
	}
	return aturans, nil
}

func GetAturanByUID(uid string) (Aturan, error) {
	var aturans Aturan
	err := database.DB.Where("aturan_uid = ?", uid).First(&aturans).Error
	if err != nil {
		return Aturan{}, err
	}
	return aturans, nil
}

func (a *Aturan) UpdateAturan(uid string) error {
	err := database.DB.Where("aturan_uid = ?", a.AturanID).Updates(a).Error
	return err
}

func DeleteAturan(uid string) error {
	err := database.DB.Where("aturan_uid = ?", uid).Delete(&Aturan{}).Error
	return err
}
