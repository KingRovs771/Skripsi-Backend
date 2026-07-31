package models

import "time"

// StudentConsent mencatat persetujuan eksplisit siswa terhadap penggunaan data
// kesehatan mental mereka sesuai UU PDP No. 27/2022 Pasal 4 huruf b.
//
// consent_version memungkinkan sistem meminta persetujuan ulang jika teks
// kebijakan privasi diperbarui di masa mendatang.
type StudentConsent struct {
	ConsentId      int64      `gorm:"primaryKey;autoIncrement"       json:"consent_id"`
	ConsentUID     string     `gorm:"type:varchar(255);uniqueIndex"  json:"consent_uid"`
	StudentUID     string     `gorm:"type:varchar(255);index"        json:"student_uid"`
	ConsentVersion string     `gorm:"type:varchar(20)"               json:"consent_version"` // e.g. "v1"
	IsAgreed       bool       `gorm:"default:false"                  json:"is_agreed"`
	AgreedAt       *time.Time `gorm:"type:timestamp"                 json:"agreed_at"`
	IPAddress      string     `gorm:"type:varchar(50)"               json:"ip_address"`
	CreatedAt      time.Time  `gorm:"type:timestamp"                 json:"created_at"`
}
