package models

import (
	"Skripsi-Backend/database"
	"strings"
	"time"

	"github.com/google/uuid"
	"html"
	"gorm.io/gorm"
)

type Pertanyaan struct {
	PertanyaanId       int       `gorm:"primaryKey;uniqueIndex" json:"pertanyaan_id"`
	PertanyaanUID      string    `gorm:"varchar(255)" json:"pertanyaan_uid"`
	KodePertanyaan     string    `gorm:"varchar(255)" json:"kode_pertanyaan"`
	KategoriPertanyaan string    `gorm:"type:varchar" json:"kategori_pertanyaan"`
	Pertanyaan         string    `gorm:"text" json:"pertanyaan"`
	Bobot              float64   `gorm:"type:decimal(10,2)" json:"bobot"`
	CreatedAt          time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt           time.Time `gorm:"type:timestamp" json:"update_at"`
}

func (a *Pertanyaan) BeforeCreate(db *gorm.DB) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	a.PertanyaanUID = uid.String()

	a.KodePertanyaan = html.EscapeString(strings.TrimSpace(a.KodePertanyaan))
	a.KategoriPertanyaan = html.EscapeString(strings.TrimSpace(a.KategoriPertanyaan))
	a.Pertanyaan = html.EscapeString(strings.TrimSpace(a.Pertanyaan))
	a.CreatedAt = time.Now()

	return nil
}

func (a *Pertanyaan) BeforeUpdate(db *gorm.DB) error {

	a.KodePertanyaan = html.EscapeString(strings.TrimSpace(a.KodePertanyaan))
	a.KategoriPertanyaan = html.EscapeString(strings.TrimSpace(a.KategoriPertanyaan))
	a.Pertanyaan = html.EscapeString(strings.TrimSpace(a.Pertanyaan))
	a.UpdateAt = time.Now()
	return nil
}

func (a *Pertanyaan) SavePertanyaan() (*Pertanyaan, error) {
	err := database.DB.Create(a).Error
	if err != nil {
		return nil, err
	}
	return a, nil
}

func GetAllPertanyaan() ([]Pertanyaan, error) {
	var quest []Pertanyaan
	err := database.DB.Find(&quest).Error
	if err != nil {
		return quest, err
	}
	return quest, nil
}

func GetPertanyaanByUID(uid string) (Pertanyaan, error) {
	var quest Pertanyaan
	err := database.DB.Where("pertanyaan_uid = ?", uid).First(&quest).Error
	if err != nil {
		return Pertanyaan{}, err
	}
	return quest, nil
}

func (a *Pertanyaan) UpdatePertanyaan(uid string) error {
	err := database.DB.Where("pertanyaan_uid = ?", uid).Updates(a).Error
	return err
}

func DeletePertanyaan(uid string) error {
	err := database.DB.Where("pertanyaan_uid = ?", uid).Delete(&Pertanyaan{}).Error
	return err
}
