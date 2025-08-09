package models

import "time"

type JawabanStudents struct {
	JawabanId    int64     `gorm:"primaryKey" json:"jawaban_id"`
	StudentUID   int64     `gorm:"type:int" json:"student_id"`
	KuisionerId  int64     `gorm:"type:int" json:"kuisioner_id"`
	NilaiJawaban int64     `gorm:"type:int" json:"nilai_jawaban"`
	TestSesiId   int64     `gorm:"type:varchar" json:"test_sesi_id"`
	JawabanAt    time.Time `gorm:"type:timestamp" json:"jawaban_at"`
}
