package models

import (
	"Skripsi-Backend/database"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

type Students struct {
	StudentsId  int64     `gorm:"primaryKey;uniqueIndex" json:"students_id"`
	StudentsUID string    `gorm:"type:varchar(255)" json:"students_uid"`
	RoleUID     string    `gorm:"type:varchar(255)" json:"role_uid"`
	NISN        string    `gorm:"type:varchar(20)" json:"nisn"`
	NamaLengkap string    `gorm:"type:varchar(90)" json:"nama_lengkap"`
	NPSN        string    `gorm:"type:varchar(30)" json:"npsn"`
	NamaSekolah string    `gorm:"->;column:nama_sekolah" json:"nama_sekolah"`
	Kelas       string    `gorm:"type:varchar(50)" json:"kelas"`
	NoHp        string    `gorm:"type:varchar(50)" json:"no_hp"`
	Alamat      string    `gorm:"type:text" json:"alamat"`
	Email       string     `gorm:"type:varchar(100)" json:"email"`
	Password    string     `gorm:"type:varchar(255)" json:"password"`
	StatusAkun  string     `gorm:"type:varchar(20);default:'AKTIF'" json:"status_akun"`      // "AKTIF", "LULUS", "PINDAH"
	TanggalNonaktif *time.Time `gorm:"type:timestamp" json:"tanggal_nonaktif"`
	CreatedAt   time.Time  `gorm:"type:timestamp" json:"created_at"`
	UpdateAt    time.Time  `gorm:"type:timestamp" json:"update_at"`
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func (u *Students) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func (u *Students) BeforeCreate(*gorm.DB) error {
	//Create UUID
	UniqueId, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	u.StudentsUID = UniqueId.String()

	//Hash Password
	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword

	u.NISN = html.EscapeString(strings.TrimSpace(u.NISN))
	u.NamaLengkap = html.EscapeString(strings.TrimSpace(u.NamaLengkap))
	u.NoHp = html.EscapeString(strings.TrimSpace(u.NoHp))
	u.Alamat = html.EscapeString(strings.TrimSpace(u.Alamat))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	if u.StatusAkun == "" {
		u.StatusAkun = "AKTIF"
	}
	u.CreatedAt = time.Now()

	return nil
}

func (u *Students) BeforeUpdate(*gorm.DB) error {

	u.NISN = html.EscapeString(strings.TrimSpace(u.NISN))
	u.NamaLengkap = html.EscapeString(strings.TrimSpace(u.NamaLengkap))
	u.NoHp = html.EscapeString(strings.TrimSpace(u.NoHp))
	u.Alamat = html.EscapeString(strings.TrimSpace(u.Alamat))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	u.UpdateAt = time.Now()

	return nil
}

func (u *Students) Save() (*Students, error) {
	err := database.DB.Create(&u).Error
	if err != nil {
		return &Students{}, err
	}
	return u, nil
}

func GetAllStudents() ([]Students, error) {
	var students []Students

	err := database.DB.Table("students").
		Select("students.*, sekolahs.nama_sekolah").
		Joins("left join sekolahs on sekolahs.npsn::text = students.npsn").
		Scan(&students).Error

	if err != nil {
		return []Students{}, err
	}
	return students, nil
}

func GetStudentsById(uid string) (Students, error) {
	var students Students
	err := database.DB.Where("students_uid = ?", uid).First(&students).Error
	if err != nil {
		return Students{}, err
	}
	return students, nil
}

func (u *Students) UpdateStudents(uid string) error {
	err := database.DB.Where("students_uid = ?", uid).First(&u).Error
	if err != nil {
		return err
	}
	return err
}

func DeleteStudents(uid string) error {
	err := database.DB.Where("students_uid = ?", uid).Delete(&Students{}).Error
	return err
}

func FindUserByEmail(email string) (Students, error) {
	var user Students
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return Students{}, err
	}
	return user, nil
}
func FindUserByID(StudentsUID string) (Students, error) {
	var student Students

	err := database.DB.Where("students_uid = ?", StudentsUID).First(&student).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Students{}, errors.New("user tidak ditemukan")
		}
		return Students{}, err
	}
	return student, nil
}
