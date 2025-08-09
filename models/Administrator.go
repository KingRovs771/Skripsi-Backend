package models

type Administrator struct {
	AdminId     int64  `gorm:"primaryKey" json:"admin_id"`
	AdminUID    string `gorm:"type:varchar(255)" json:"admin_uid"`
	NamaLengkap string `gorm:"type:varchar(90)" json:"nama_lengkap"`
	Phone       string `gorm:"type:varchar(20)" json:"phone"`
	Email       string `gorm:"type:varchar(90)" json:"email"`
	Alamat      string `gorm:"type:text" json:"alamat"`
	Username    string `gorm:"type:varchar(90)" json:"username"`
	Password    string `gorm:"type:varchar(90)" json:"password"`
	PhotoFile   []byte `gorm:"type:longblob;not null" json:"face_data"`
}
