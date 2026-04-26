package models

type TestAnswer struct {
	TestAnswer     string `gorm:"primary_key;uniqueIndex" json:"test_answer"`
	TestSessionUID string `gorm:"varchar(255);index" json:"test_session_id"`
	KodePertanyaan string `gorm:"type:varchar(30)" json:"kode_pertanyaan"`
	NilaiJawaban   int64  `gorm:"bigint" json:"nilai_jawaban"`
}
