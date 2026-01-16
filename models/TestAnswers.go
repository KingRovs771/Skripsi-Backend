package models

type TestAnswer struct {
	TestAnswer     string `gorm:"primary_key;uniqueIndex" json:"test_answer"`
	TestSessionUID string `gorm:"varchar(255)" json:"test_session_id"`
	PertanyaanUID  string `gorm:"varchar(255)" json:"pertanyaan_uid"`
	NilaiJawaban   int64  `gorm:"bigint" json:"nilai_jawaban"`
}
