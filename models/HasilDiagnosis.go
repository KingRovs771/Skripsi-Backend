package models

import "time"

type HasilDiagnosis struct {
	ResultId             int       `gorm:"primary_key;uniqueIndex" json:"result_id"`
	ResultUID            string    `gorm:"varchar(255)" json:"result_uid"`
	SessionTestUID       string    `gorm:"varchar(255)" json:"session_test_uid"`
	NNPredictionPenyakit string    `gorm:"varchar(255)" json:"nn_prediction_penyakit"`
	NNConfidenceScore    float64   `gorm:"type:decimal(10,2)" json:"nn_confidence_score"`
	BCVerificationStatus string    `gorm:"varchar(255)" json:"bc_verification_status"`
	FinalPenyakit        string    `gorm:"varchar(255)" json:"final_penyakit"`
	CreatedAt            time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt             time.Time `gorm:"type:timestamp" json:"update_at"`
}
