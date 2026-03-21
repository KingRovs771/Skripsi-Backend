package models

import (
	"Skripsi-Backend/database"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

type Pakar struct {
	PakarID        int64     `gorm:"primaryKey;uniqueIndex" json:"pakar_id"`
	PakarUID       string    `gorm:"type:varchar(255)" json:"pakar_uid"`
	RoleUID        string    `gorm:"type:varchar(255)" json:"role_uid"`
	NomorSIP       string    `gorm:"type:varchar(90)" json:"nomor_sip"`
	NamaLengkap    string    `gorm:"type:varchar(90)" json:"nama_lengkap"`
	JenisSpesialis string    `gorm:"type:varchar" json:"jenis_spesialis"`
	Phone          string    `gorm:"type:varchar(20)" json:"phone"`
	Email          string    `gorm:"type:varchar(90)" json:"email"`
	Alamat         string    `gorm:"type:text" json:"alamat"`
	Password       string    `gorm:"type:varchar(90)" json:"password"`
	PhotoFile      []byte    `gorm:"type:bytea;not null" json:"face_data"`
	CreatedAt      time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt       time.Time `gorm:"type:timestamp" json:"update_at"`
}

func (u *Pakar) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func (p *Pakar) BeforeSave(*gorm.DB) error {
	uid, err := uuid.NewRandom()
	p.PakarUID = uid.String()
	if err != nil {
		return err
	}
	p.NomorSIP = html.EscapeString(strings.TrimSpace(p.NomorSIP))
	p.NamaLengkap = html.EscapeString(strings.TrimSpace(p.NamaLengkap))
	p.JenisSpesialis = html.EscapeString(strings.TrimSpace(p.JenisSpesialis))
	p.Phone = html.EscapeString(strings.TrimSpace(p.Phone))
	p.Email = html.EscapeString(strings.TrimSpace(p.Email))
	p.Alamat = html.EscapeString(strings.TrimSpace(p.Alamat))

	return nil
}

func (p *Pakar) BeforeUpdate(*gorm.DB) error {

	p.NomorSIP = html.EscapeString(strings.TrimSpace(p.NomorSIP))
	p.NamaLengkap = html.EscapeString(strings.TrimSpace(p.NamaLengkap))
	p.JenisSpesialis = html.EscapeString(strings.TrimSpace(p.JenisSpesialis))
	p.Phone = html.EscapeString(strings.TrimSpace(p.Phone))
	p.Email = html.EscapeString(strings.TrimSpace(p.Email))
	p.Alamat = html.EscapeString(strings.TrimSpace(p.Alamat))
	p.UpdateAt = time.Now()

	return nil
}

func (p *Pakar) SaveUsersPakar() (*Pakar, error) {
	err := database.DB.Create(&p).Error
	if err != nil {
		return &Pakar{}, err
	}
	return p, nil
}

func GetAllPakar() ([]Pakar, error) {
	var PakarList []Pakar

	err := database.DB.Find(&PakarList).Error
	if err != nil {
		return []Pakar{}, err
	}
	return PakarList, nil
}

func GetPakarByUID(uid string) (Pakar, error) {
	var pakar Pakar
	err := database.DB.Where("pakar_uid = ?", uid).First(&pakar).Error
	if err != nil {
		return Pakar{}, err
	}
	return pakar, nil
}

func (p *Pakar) UpdatePakar(uid string) error {
	err := database.DB.Where("pakar_uid = ?", uid).Updates(p).Error
	return err
}

func DeletePakar(uid string) error {
	err := database.DB.Where("pakar_uid = ?", uid).Delete(&Pakar{}).Error
	return err
}

func FindUserByEmailPakars(email string) (Pakar, error) {
	var user Pakar
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return Pakar{}, err
	}
	return user, nil
}
