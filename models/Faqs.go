package models

import "time"

type Faqs struct {
	FaqsId        int       `gorm:"primaryKey;uniqueIndex" json:"faqs_id"`
	FaqsUID       string    `gorm:"varchar(255)" json:"faqs_uid"`
	ClientUID     *string   `gorm:"varchar(255)" json:"client_uid"` //Ke Pakar / GuruBK
	UserUID       string    `gorm:"varchar(255)" json:"user_uid"`   // Yang Upload
	FaqPertanyaan string    `gorm:"varchar(255)" json:"faq_pertanyaan"`
	CreatedAt     time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt      time.Time `gorm:"type:timestamp" json:"update_at"`
}
