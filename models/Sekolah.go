package models

import (
	"Skripsi-Backend/database"
	"errors"
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

func (s *Sekolah) SaveSekolah() (*Sekolah, error) {
	var count int64
	database.DB.Model(&Sekolah{}).Where("npsn = ?", s.NPSN).Count(&count)

	if count > 0 {
		return nil, errors.New("NPSN tersebut sudah terdaftar di sistem")
	}

	err := database.DB.Create(&s).Error
	if err != nil {
		return nil, err
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

func SearchSekolah(query string) ([]Sekolah, error) {
	var sekolahs []Sekolah

	searchQuery := "%" + query + "%"

	err := database.DB.Where("nama_sekolah ILIKE ? OR CAST(npsn AS TEXT) LIKE ?", searchQuery, searchQuery).
		Limit(10).
		Find(&sekolahs).Error

	return sekolahs, err

}
