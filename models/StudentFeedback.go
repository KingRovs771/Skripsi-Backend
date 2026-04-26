package models

import "time"

type StudentFeedback struct {
	FeedbackId     int64     `gorm:"primaryKey;autoIncrement" json:"feedback_id"`
	TestSessionUID string    `gorm:"type:varchar(255);index" json:"test_session_uid"`
	CeritaSiswa    string    `gorm:"type:text" json:"cerita_siswa"`
	CreatedAt      time.Time `gorm:"type:timestamp" json:"created_at"`
}
