package models

import (
	"Skripsi-Backend/database"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/html"
	"gorm.io/gorm"
	"strings"
)

type Students struct {
	gorm.Model
	StudentsId        int64  `gorm:"primaryKey" json:"students_id"`
	StudentsUID       string `gorm:"type:varchar(90)" json:"students_uid"`
	NISN              string `gorm:"type:varchar(20)" json:"nisn"`
	NamaLengkap       string `gorm:"type:varchar(90)" json:"nama_lengkap"`
	NamaInisial       string `gorm:"type:varchar(90)" json:"nama_inisial"`
	JenjangPendidikan string `gorm:"type:varchar" json:"jenjang_pendidikan"`
	Kelas             int64  `gorm:"type:int" json:"kelas"`
	Email             string `gorm:"type:varchar(100)" json:"email"`
	Username          string `gorm:"type:varchar(90)" json:"username"`
	Password          string `gorm:"type:varchar(255)" json:"password"`
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // 14 adalah cost factor
	return string(bytes), err
}

func (u *Students) BeforeSave(*gorm.DB) error {
	// 1. Hash password user.
	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword

	u.Username = html.EscapeString(strings.TrimSpace(u.Username))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))

	return nil
}

func (u *Students) Save() (*Students, error) {
	err := database.DB.Create(&u).Error
	if err != nil {
		return &Students{}, err
	}
	return u, nil
}

func (u *Students) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func FindUserByEmail(email string) (Students, error) {
	var user Students
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return Students{}, err
	}
	return user, nil
}
func FindUserByID(StudentsUID int64) (Students, error) {
	var user Students
	err := database.DB.First(&user, StudentsUID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Students{}, errors.New("user tidak ditemukan")
		}
		return Students{}, err
	}
	return user, nil
}
