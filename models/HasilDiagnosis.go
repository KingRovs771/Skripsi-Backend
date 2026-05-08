package models

import "time"

type HasilDiagnosis struct {
	ResultId              int       `gorm:"primary_key;autoIncrement" json:"result_id"`
	ResultUID             string    `gorm:"varchar(255);uniqueIndex" json:"result_uid"`
	SessionTestUID        string    `gorm:"varchar(255);index" json:"session_test_uid"`
	NNDepresiPrediksi     string    `gorm:"type:varchar(30)" json:"nn_depresi_prediksi"`
	NNDepresiConfidence   float64   `gorm:"type:decimal(5,2)" json:"nn_depresi_confidence"`
	FinalDepresiPenyakit  string    `gorm:"type:varchar(30)" json:"final_depresi_penyakit"`
	StatusValidasiDepresi string    `gorm:"type:varchar(50)" json:"status_validasi_depresi"` // CONFIRMED / ADJUSTED
	NNCemasPrediksi       string    `gorm:"type:varchar(30)" json:"nn_cemas_prediksi"`
	NNCemasConfidence     float64   `gorm:"type:decimal(5,2)" json:"nn_cemas_confidence"`
	FinalCemasPenyakit    string    `gorm:"type:varchar(30)" json:"final_cemas_penyakit"`
	StatusValidasiCemas   string    `gorm:"type:varchar(50)" json:"status_validasi_cemas"` // CONFIRMED / ADJUSTED
	TinjauanBK            string    `gorm:"type:varchar(50);default:'MENUNGGU'" json:"tinjauan_bk"`
	IsVisibleToStudent    bool      `gorm:"default:false" json:"is_visible_to_student"`
	ReviewedByGurubk      bool      `gorm:"default:false" json:"reviewed_by_gurubk"`
	Rekomendasi           string    `gorm:"type:text" json:"rekomendasi"`
	CreatedAt             time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt              time.Time `gorm:"type:timestamp" json:"update_at"`
}
