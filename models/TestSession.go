package models

import "time"

type TestSession struct {
	TestSessionId  string    `gorm:"primary_key;uniqueIndex" json:"test_session_id"`
	UserUID        string    `gorm:"varchar(255)" json:"user_uid"`
	TotalScorephq9 int64     `gorm:"bigint" json:"total_scorephq9"`
	TotalScoregad7 int64     `gorm:"bigint" json:"total_scoregad7"`
	Status         string    `gorm:"type:varchar(50);default:'SELESAI'" json:"status"`
	CreatedAt      time.Time `gorm:"type:timestamp" json:"created_at"`
}
