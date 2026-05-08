package models

type TestAnswer struct {
	TestAnswerId   string `gorm:"primary_key;uniqueIndex" json:"test_answer_id"`
	TestSessionUID string `gorm:"varchar(255);index" json:"test_session_uid"`
	KodePertanyaan string `gorm:"type:varchar(30)" json:"kode_pertanyaan"`
	NilaiJawaban   int64  `gorm:"bigint" json:"nilai_jawaban"`
}
