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

type Teachers struct {
	TeachersID  int64     `gorm:"primaryKey;uniqueIndex" json:"teachers_id"`
	TeachersUID string    `gorm:"type:varchar(255)" json:"teachers_uid"`
	RoleUID     string    `gorm:"type:varchar(255)" json:"role_uid"`
	NIP         string    `gorm:"type:varchar(20)" json:"nip"`
	NamaLengkap string    `gorm:"type:varchar(90)" json:"nama_lengkap"`
	NPSN        string    `gorm:"type:varchar(30)" json:"npsn"`
	NamaSekolah string    `gorm:"->;column:nama_sekolah" json:"nama_sekolah"`
	Phone       string    `gorm:"type:varchar(20)" json:"phone"`
	Alamat      string    `gorm:"type:text" json:"alamat"`
	Email       string    `gorm:"type:varchar(90)" json:"email"`
	Password    string    `gorm:"type:varchar(90)" json:"password"`
	PhotoFile   []byte    `gorm:"type:bytea" json:"face_data"`
	CreatedAt   time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt    time.Time `gorm:"type:timestamp" json:"update_at"`
}

func (u *Teachers) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func hashPasswordTeachers(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (t *Teachers) BeforeCreate(tx *gorm.DB) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	t.TeachersUID = uid.String()

	hashedPassword, err := hashPasswordTeachers(t.Password)
	if err != nil {
		return err
	}
	t.Password = hashedPassword

	t.NamaLengkap = html.EscapeString(strings.TrimSpace(t.NamaLengkap))
	t.Email = html.EscapeString(strings.TrimSpace(t.Email))
	t.NIP = html.EscapeString(strings.TrimSpace(t.NIP))
	t.Alamat = html.EscapeString(strings.TrimSpace(t.Alamat))

	return nil
}

func (t *Teachers) BeforeUpdate(tx *gorm.DB) error {
	t.NamaLengkap = html.EscapeString(strings.TrimSpace(t.NamaLengkap))
	t.Email = html.EscapeString(strings.TrimSpace(t.Email))
	t.NIP = html.EscapeString(strings.TrimSpace(t.NIP))
	t.Alamat = html.EscapeString(strings.TrimSpace(t.Alamat))
	t.UpdateAt = time.Now()
	return nil
}

// SaveTeachers menyimpan data guru baru
func (t *Teachers) SaveTeachers() (*Teachers, error) {
	err := database.DB.Create(&t).Error
	if err != nil {
		return &Teachers{}, err
	}
	return t, nil
}

// GetAllTeachers mengambil semua data guru
func GetAllTeachers() ([]Teachers, error) {
	var teachersList []Teachers
	err := database.DB.Table("teachers").
		Select("teachers.teachers_uid, teachers.n_ip, teachers.nama_lengkap, teachers.npsn,teachers.email, teachers.phone, sekolahs.nama_sekolah").
		Joins("left join sekolahs on sekolahs.npsn::text = teachers.npsn").
		Scan(&teachersList).Error
	if err != nil {
		return nil, err
	}
	return teachersList, nil
}

// GetTeachersByUID mengambil data guru berdasarkan UID
func GetTeachersByUID(uid string) (Teachers, error) {
	var teacher Teachers
	err := database.DB.Where("teachers_uid = ?", uid).First(&teacher).Error
	if err != nil {
		return Teachers{}, err
	}
	return teacher, nil
}

// UpdateTeachers memperbarui data guru
func (t *Teachers) UpdateTeachers(uid string) error {
	err := database.DB.Where("teachers_uid = ?", uid).Updates(t).Error
	return err
}

// DeleteTeachers menghapus data guru berdasarkan UID
func DeleteTeachers(uid string) error {
	err := database.DB.Where("teachers_uid = ?", uid).Delete(&Teachers{}).Error
	return err
}

func FindUserByEmailTeachers(email string) (Teachers, error) {
	var user Teachers
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return Teachers{}, err
	}
	return user, nil
}
