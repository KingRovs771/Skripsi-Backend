package models

import "time"

type Teachers struct {
	TeachersID  int64     `gorm:"primaryKey;uniqueIndex" json:"teachers_id"`
	TeachersUID string    `gorm:"type:varchar(255)" json:"teachers_uid"`
	RoleUID     int64     `gorm:"type:int" json:"role_uid"`
	NIP         string    `gorm:"type:varchar(20)" json:"nip"`
	NamaLengkap string    `gorm:"type:varchar(90)" json:"nama_lengkap"`
	NPSN        string    `gorm:"type:int" json:"npsn"`
	Phone       string    `gorm:"type:varchar(20)" json:"phone"`
	Email       string    `gorm:"type:varchar(90)" json:"email"`
	Alamat      string    `gorm:"type:text" json:"alamat"`
	Username    string    `gorm:"type:varchar(90)" json:"username"`
	Password    string    `gorm:"type:varchar(90)" json:"password"`
	PhotoFile   []byte    `gorm:"type:bytea;not null" json:"face_data"`
	CreatedAt   time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt    time.Time `gorm:"type:timestamp" json:"update_at"`
}
