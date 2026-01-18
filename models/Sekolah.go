package models

import (
	"Skripsi-Backend/database"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

type Sekolah struct {
	SekolahId     int64     `gorm:"primaryKey;uniqueIndex" json:"sekolah_id"`
	SekolahUID    string    `gorm:"type:varchar(255)" json:"sekolah_uid"`
	NPSN          int64     `gorm:"type:int" json:"npsn"`
	NamaSekolah   string    `gorm:"type:varchar(100)" json:"nama_sekolah"`
	Jenjang       string    `gorm:"type:varchar(20)" json:"jenjang"`
	AlamatSekolah string    `gorm:"type:text" json:"alamat_sekolah"`
	CreatedAt     time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt      time.Time `gorm:"type:timestamp" json:"update_at"`
}

func (s *Sekolah) BeforeCreate(tx *gorm.DB) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	s.SekolahUID = uid.String()

	s.NamaSekolah = html.EscapeString(strings.TrimSpace(s.NamaSekolah))
	s.Jenjang = html.EscapeString(strings.TrimSpace(s.Jenjang))
	s.AlamatSekolah = html.EscapeString(strings.TrimSpace(s.AlamatSekolah))
	return nil
}

func (s *Sekolah) BeforeUpdate(tx *gorm.DB) error {
	s.NamaSekolah = html.EscapeString(strings.TrimSpace(s.NamaSekolah))
	s.Jenjang = html.EscapeString(strings.TrimSpace(s.Jenjang))
	s.AlamatSekolah = html.EscapeString(strings.TrimSpace(s.AlamatSekolah))
	s.UpdateAt = time.Now()
	return nil
}

func GetAllSekolah() ([]Sekolah, error) {
	var SekolahList []Sekolah
	err := database.DB.Find(&SekolahList).Error
	if err != nil {
		return SekolahList, err
	}
	return SekolahList, nil
}

func (s *Sekolah) SaveSekolah() (*Sekolah, error) {
	var Sekolah Sekolah
	err := database.DB.Create(&Sekolah).Error
	if err != nil {
		return &Sekolah, err
	}
	return s, nil
}

func GetSekolahByUID(uid string) (Sekolah, error) {
	var sekolah Sekolah
	err := database.DB.Where("sekolah_uid = ?", uid).First(&sekolah).Error
	if err != nil {
		return Sekolah{}, err
	}
	return sekolah, nil
}

func (s *Sekolah) UpdateSekolah(uid string) error {
	err := database.DB.Where("sekolah_uid = ?", uid).Updates(s).Error
	return err
}

func DeleteSekolah(uid string) error {
	err := database.DB.Where("sekolah_uid = ?", uid).Delete(&Sekolah{}).Error
	return err
}
