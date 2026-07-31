package utils

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"log"
	"time"
)

// StartRetentionScheduler memulai background worker untuk membersihkan/anonimisasi
// data siswa yang telah melewati masa retensi (2 tahun sejak LULUS/PINDAH).
func StartRetentionScheduler() {
	// Jalankan pembersihan saat startup (asinkron agar tidak memblokir server boot)
	go func() {
		// Tunggu sebentar sampai database terkoneksi sepenuhnya
		time.Sleep(5 * time.Second)
		RunRetentionCleanup()
	}()

	// Jadwalkan untuk berjalan setiap 24 jam
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			RunRetentionCleanup()
		}
	}()
}

// RunRetentionCleanup mencari siswa yang berstatus LULUS/PINDAH dengan tanggal_nonaktif
// lebih dari 2 tahun yang lalu, lalu melakukan anonimisasi terhadap data mereka.
func RunRetentionCleanup() {
	log.Println("[Retention Job] Memulai pengecekan kebijakan retensi data...")

	// Periode retensi: 2 tahun
	retentionPeriod := 2 * 365 * 24 * time.Hour
	cutoff := time.Now().Add(-retentionPeriod)

	var expiredStudents []models.Students
	// Query siswa yang statusnya bukan AKTIF dan tanggal_nonaktif lebih tua dari 2 tahun
	err := database.DB.
		Where("status_akun != 'AKTIF' AND tanggal_nonaktif IS NOT NULL AND tanggal_nonaktif < ?", cutoff).
		Find(&expiredStudents).Error

	if err != nil {
		log.Printf("[Retention Job] Gagal query siswa kadaluarsa: %v\n", err)
		return
	}

	if len(expiredStudents) == 0 {
		log.Println("[Retention Job] Tidak ada siswa yang melewati batas retensi data.")
		return
	}

	log.Printf("[Retention Job] Menemukan %d siswa kadaluarsa. Memulai anonimisasi...\n", len(expiredStudents))

	for _, student := range expiredStudents {
		tx := database.DB.Begin()

		// 1. Anonimkan data utama siswa
		anonNISN := "ANONIM-" + student.StudentsUID
		anonEmail := "anonim-" + student.StudentsUID + "@sindas.id"

		err = tx.Model(&student).Updates(map[string]interface{}{
			"nama_lengkap": "Siswa Anonim",
			"nisn":         anonNISN,
			"email":        anonEmail,
			"no_hp":        "",
			"alamat":       "",
			"kelas":        "",
		}).Error

		if err != nil {
			tx.Rollback()
			log.Printf("[Retention Job] Gagal melakukan anonimisasi profil siswa %s: %v\n", student.StudentsUID, err)
			continue
		}

		// 2. Cari semua test_sessions milik siswa tersebut
		var sessions []models.TestSession
		err = tx.Where("user_uid = ?", student.StudentsUID).Find(&sessions).Error
		if err != nil {
			tx.Rollback()
			log.Printf("[Retention Job] Gagal mengambil sesi tes siswa %s: %v\n", student.StudentsUID, err)
			continue
		}

		// 3. Anonimkan narasi/cerita_siswa pada tabel student_feedbacks untuk setiap sesi tes
		for _, session := range sessions {
			err = tx.Model(&models.StudentFeedback{}).
				Where("test_session_uid = ?", session.TestSessionId).
				Update("cerita_siswa", "[Dianonimkan sesuai kebijakan retensi data]").Error
			if err != nil {
				log.Printf("[Retention Job] Gagal anonimisasi cerita_siswa untuk sesi %s: %v\n", session.TestSessionId, err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("[Retention Job] Gagal commit transaksi anonimisasi siswa %s: %v\n", student.StudentsUID, err)
		} else {
			log.Printf("[Retention Job] Berhasil anonimisasi siswa %s.\n", student.StudentsUID)
		}
	}

	log.Println("[Retention Job] Pengecekan kebijakan retensi data selesai.")
}
