package models

import "time"

type HasilDiagnosis struct {
	ResultId         int64     `gorm:"primaryKey;uniqueIndex" json:"result_id"`
	StudentId        int64     `gorm:"type:int" json:"student_id"`
	TestSessionId    int64     `gorm:"type:varchar" json:"test_session_id"`
	Phq9Score        int64     `gorm:"type:int" json:"phq9_score"`
	Gad7Score        int64     `gorm:"type:int" json:"gad7_score"`
	StressScore      int64     `gorm:"type:int" json:"stress_score"`
	FinalDiagnosis   string    `gorm:"type:varchar(255)" json:"final_diagnosis"`
	ConfidenceLevel  float64   `gorm:"type:float" json:"confidence_level"`
	IsVerfiedByRules bool      `gorm:"type:bool" json:"is_verfied_by_rules"`
	CreatedAt        time.Time `gorm:"type:timestamp" json:"created_at"`
}
