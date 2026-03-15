package models

import (
	"Skripsi-Backend/database"
	"errors"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Administrator struct {
	AdminId     int64     `gorm:"primaryKey;uniqueIndex" json:"admin_id"`
	AdminUID    string    `gorm:"type:varchar(255)" json:"admin_uid"`
	RoleUID     string    `gorm:"type:varchar" json:"role_uid"`
	NamaLengkap string    `gorm:"type:varchar(90)" json:"nama_lengkap"`
	Phone       string    `gorm:"type:varchar(20)" json:"phone"`
	Email       string    `gorm:"type:varchar(90)" json:"email"`
	Alamat      string    `gorm:"type:text" json:"alamat"`
	Password    string    `gorm:"type:varchar(255)" json:"password"`
	PhotoFile   []byte    `gorm:"type:bytea" json:"face_data"`
	CreatedAt   time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt    time.Time `gorm:"type:timestamp" json:"update_at"`

	RoleName string `gorm:"->;column:role_name"`
}

func hashPasswordAdministrator(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (u *Administrator) BeforeSaveAdministrator(*gorm.DB) error {
	//Create UUID
	UniqueId, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	u.AdminUID = UniqueId.String()

	//Hash Password
	hashedPassword, err := hashPasswordAdministrator(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword

	u.Email = html.EscapeString(strings.TrimSpace(u.Email))

	return nil
}

func (u *Administrator) SaveAdministrator() (*Administrator, error) {
	err := database.DB.Create(&u).Error
	if err != nil {
		return &Administrator{}, err
	}
	return u, nil
}

func (u *Administrator) ValidatePasswordAdministrator(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func FindUserByEmailAdministrator(email string) (Administrator, error) {
	var user Administrator
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return Administrator{}, err
	}
	return user, nil
}

func FindUserByIDAdministrator(AdminUID string) (Administrator, error) {
	var admin Administrator
	// Gunakan .Where() untuk mencari di kolom spesifik "students_uid"
	err := database.DB.Where("admin_uid = ?", AdminUID).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Administrator{}, errors.New("user tidak ditemukan")
		}
		return Administrator{}, err
	}
	return admin, nil
}
