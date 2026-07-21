package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Faqs struct {
	FaqsId        int       `gorm:"primaryKey;uniqueIndex" json:"faqs_id"`
	FaqsUID       string    `gorm:"type:varchar(255);uniqueIndex" json:"faqs_uid"`
	UserUID       string    `gorm:"type:varchar(255);index" json:"user_uid"`
	Tujuan        string    `gorm:"type:varchar(20)" json:"tujuan"`
	FaqPertanyaan string    `gorm:"type:text" json:"faq_pertanyaan"`
	FaqJawaban    string    `gorm:"type:text" json:"faq_jawaban"`
	Status        string    `gorm:"type:varchar(20);default:'Menunggu'" json:"status"`
	RepliedBy     string    `gorm:"type:varchar(255)" json:"replied_by"`
	CreatedAt     time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt      time.Time `gorm:"type:timestamp" json:"update_at"`
}

func (f *Faqs) BeforeCreate(tx *gorm.DB) (err error) {
	f.FaqsUID = uuid.New().String()
	f.CreatedAt = time.Now()
	f.UpdateAt = time.Now()
	if f.Status == "" {
		f.Status = "Menunggu"
	}
	return
}
