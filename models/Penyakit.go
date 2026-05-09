package models

import (
	"Skripsi-Backend/database"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

type Penyakit struct {
	PenyakitID      int64     `gorm:"primaryKey;autoIncrement" json:"penyakit_id"`
	PenyakitUID     string    `gorm:"type:varchar;uniqueIndex" json:"penyakit_uid"`
	KodePenyakit    string    `gorm:"type:varchar(30);unique" json:"kode_penyakit"`
	NamaPenyakit    string    `gorm:"type:varchar" json:"nama_penyakit"`
	KodeTurunan     string    `gorm:"type:varchar(30)" json:"kode_turunan"`
	Description     string    `gorm:"type:text" json:"description"`
	SaranPenanganan string    `gorm:"type:text" json:"saran_penanganan"`
	CreatedAt       time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt        time.Time `gorm:"type:timestamp" json:"update_at"`

	DaftarAturan []Aturan `gorm:"foreignKey:KodePenyakit;references:KodePenyakit"`
}

func (p *Penyakit) BeforeSave(db *gorm.DB) error {
	if p.PenyakitUID == "" {
		uid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		p.PenyakitUID = uid.String()
	}

	p.KodePenyakit = html.EscapeString(strings.TrimSpace(p.KodePenyakit))
	p.NamaPenyakit = html.EscapeString(strings.TrimSpace(p.NamaPenyakit))
	p.KodeTurunan = html.EscapeString(strings.TrimSpace(p.KodeTurunan))
	p.Description = html.EscapeString(strings.TrimSpace(p.Description))
	p.SaranPenanganan = html.EscapeString(strings.TrimSpace(p.SaranPenanganan))
	p.CreatedAt = time.Now()

	return nil
}

func (p *Penyakit) BeforeUpdate(db *gorm.DB) error {
	p.KodePenyakit = html.EscapeString(strings.TrimSpace(p.KodePenyakit))
	p.NamaPenyakit = html.EscapeString(strings.TrimSpace(p.NamaPenyakit))
	p.KodeTurunan = html.EscapeString(strings.TrimSpace(p.KodeTurunan))
	p.Description = html.EscapeString(strings.TrimSpace(p.Description))
	p.SaranPenanganan = html.EscapeString(strings.TrimSpace(p.SaranPenanganan))
	p.UpdateAt = time.Now()

	return nil
}

func (p *Penyakit) SavePenyakit() (*Penyakit, error) {
	err := database.DB.Create(p).Error
	if err != nil {
		return nil, err
	}
	return p, err
}

func GetAllPenyakit() ([]Penyakit, error) {
	var penyakits []Penyakit
	err := database.DB.Find(&penyakits).Error
	if err != nil {
		return penyakits, err
	}
	return penyakits, err
}

func GetPenyakitByUID(uid string) (Penyakit, error) {
	var penyakit Penyakit
	err := database.DB.Where("penyakit_uid = ?", uid).First(&penyakit).Error
	if err != nil {
		return Penyakit{}, err
	}
	return penyakit, err
}

func (p *Penyakit) UpdatePenyakit(uid string) error {
	err := database.DB.Where("penyakit_uid = ?", uid).Updates(p).Error
	return err
}

func DeletePenyakit(uid string) error {
	err := database.DB.Where("penyakit_uid=?", uid).Delete(&Penyakit{}).Error
	return err
}
