package models

type DetailDiagnosis struct {
	IdDetailHasil int `gorm:"primaryKey;uniqueIndex" json:"id_detail_hasil"`
	ResultId      int `gorm:"type:int" json:"result_id"`
	IdPertanyaan  int `gorm:"type:int" json:"id_pertanyaan"`
	JawabanId     int `gorm:"typeint" json:"jawaban_id"`
}