package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BullyReport merepresentasikan laporan kasus perundungan dari siswa.
// Catatan anonimitas: PelaporUID selalu disimpan (Opsi B / pseudo-anonymous).
// Jika IsAnonim = true, field ini TIDAK ditampilkan ke UI Guru BK,
// hanya bisa diakses Admin untuk keperluan anti-spam/audit.
type BullyReport struct {
	ReportID          int64      `gorm:"primaryKey;autoIncrement"                       json:"report_id"`
	ReportUID         string     `gorm:"type:varchar(255);uniqueIndex;not null"          json:"report_uid"`
	PelaporUID        string     `gorm:"type:varchar(255)"                               json:"pelapor_uid,omitempty"` // FK ke students.students_uid
	NPSN              string     `gorm:"type:varchar(30);not null"                       json:"npsn"`
	JenisBully        string     `gorm:"type:varchar(50);not null"                       json:"jenis_bully"`   // Verbal/Fisik/Cyberbullying/Sosial/Lainnya
	NamaTerlapor      string     `gorm:"type:varchar(100)"                               json:"nama_terlapor"`
	KelasTerlapor     string     `gorm:"type:varchar(50)"                                json:"kelas_terlapor"`
	DeskripsiKejadian string     `gorm:"type:text;not null"                              json:"deskripsi_kejadian"`
	LokasiKejadian    string     `gorm:"type:varchar(200)"                               json:"lokasi_kejadian"`
	TanggalKejadian   *time.Time `gorm:"type:date"                                       json:"tanggal_kejadian"`
	IsAnonim          bool       `gorm:"not null;default:false"                          json:"is_anonim"`
	TingkatUrgensi    string     `gorm:"type:varchar(30);not null"                       json:"tingkat_urgensi"` // Rendah/Sedang/Tinggi-Darurat
	Status            string     `gorm:"type:varchar(30);not null;default:'BARU'"        json:"status"`          // BARU/DITINDAKLANJUTI/SELESAI/DITOLAK
	DitanganiOleh     string     `gorm:"type:varchar(255)"                               json:"ditangani_oleh"`  // FK ke teachers.teachers_uid
	CatatanPenanganan string     `gorm:"type:text"                                       json:"catatan_penanganan"`
	NotifikasiTerkirim bool      `gorm:"not null;default:false"                          json:"notifikasi_terkirim"`
	CreatedAt         time.Time  `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime"                                  json:"updated_at"`
}

func (b *BullyReport) BeforeCreate(tx *gorm.DB) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	b.ReportUID = uid.String()
	return nil
}
