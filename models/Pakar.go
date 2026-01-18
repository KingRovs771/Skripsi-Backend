package models

import (
	"Skripsi-Backend/database"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Pakar struct {
	PakarID        int64     `gorm:"primaryKey;uniqueIndex" json:"pakar_id"`
	PakarUID       string    `gorm:"type:varchar(255)" json:"pakar_uid"`
	RoleUID        int64     `gorm:"type:int" json:"role_uid"`
	NomorSIP       string    `gorm:"type:varchar(20)" json:"nomor_sip"`
	NamaLengkap    string    `gorm:"type:varchar(90)" json:"nama_lengkap"`
	JenisSpesialis string    `gorm:"type:int" json:"jenis_spesialis"`
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

func FindUserByEmailPakars(email string) (Pakar, error) {
	var user Pakar
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return Pakar{}, err
	}
	return user, nil
}
