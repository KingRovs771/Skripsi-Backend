package seeder

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"log"

	"gorm.io/gorm"
)

func SeederPenyakitAturan() {
	db := database.DB

	// ─────────────────────────────────────────────────────────────────────────
	// A.1 & A.2: Update data Penyakit dengan min_skor, max_skor, dan kode_turunan
	// ─────────────────────────────────────────────────────────────────────────
	penyakitData := []struct {
		KodePenyakit    string
		NamaPenyakit    string
		MinSkor         int64
		MaxSkor         int64
		KodeTurunan     string
		SaranPenanganan string
		Description     string
	}{
		{"P05", "Depresi Berat", 20, 27, "P04", "Konsultasi ke Psikiater/Psikolog, Psikoterapi intensif, Farmakoterapi.", "Depresi kategori berat."},
		{"P04", "Depresi Sedang Berat", 15, 19, "P03", "Konsultasi ke Psikolog, Psikoterapi, pertimbangkan farmakoterapi.", "Depresi kategori sedang-berat."},
		{"P03", "Depresi Sedang", 10, 14, "P02", "Konsultasi ke Psikolog, konseling, support group.", "Depresi kategori sedang."},
		{"P02", "Depresi Ringan", 5, 9, "P01", "Self-help, olahraga teratur, konseling ringan.", "Depresi kategori ringan."},
		{"P01", "Depresi Minimal / Normal", 0, 4, "", "Pertahankan pola hidup sehat, kelola stres.", "Tidak ada gejala depresi klinis."},
		{"K04", "Kecemasan Berat", 15, 21, "K03", "Psikoterapi (CBT), relaksasi intensif, konsultasi medis.", "Kecemasan kategori berat."},
		{"K03", "Kecemasan Sedang", 10, 14, "K02", "CBT, teknik pernapasan, relaksasi otot progresif.", "Kecemasan kategori sedang."},
		{"K02", "Kecemasan Ringan", 5, 9, "K01", "Teknik relaksasi, olahraga, kurangi kafein.", "Kecemasan kategori ringan."},
		{"K01", "Kecemasan Minimal / Normal", 0, 4, "", "Pertahankan manajemen stres yang baik.", "Tidak ada gejala kecemasan klinis."},
	}

	for _, p := range penyakitData {
		var existing models.Penyakit
		err := db.Where("kode_penyakit = ?", p.KodePenyakit).First(&existing).Error
		minVal := p.MinSkor
		maxVal := p.MaxSkor
		if err == nil {
			// Update existing record
			db.Model(&existing).Updates(map[string]interface{}{
				"min_skor":     &minVal,
				"max_skor":     &maxVal,
				"kode_turunan": p.KodeTurunan,
			})
			log.Printf("✅ Penyakit %s updated", p.KodePenyakit)
		} else if err == gorm.ErrRecordNotFound {
			// Create new record
			newPenyakit := models.Penyakit{
				KodePenyakit:    p.KodePenyakit,
				NamaPenyakit:    p.NamaPenyakit,
				MinSkor:         &minVal,
				MaxSkor:         &maxVal,
				KodeTurunan:     p.KodeTurunan,
				SaranPenanganan: p.SaranPenanganan,
				Description:     p.Description,
			}
			if err := db.Create(&newPenyakit).Error; err != nil {
				log.Printf("❌ Gagal membuat Penyakit %s: %v", p.KodePenyakit, err)
			} else {
				log.Printf("✅ Penyakit %s created", p.KodePenyakit)
			}
		}
	}

	// ─────────────────────────────────────────────────────────────────────────
	// A.3: Migrasi Ulang Data Aturan (Aturans)
	// ─────────────────────────────────────────────────────────────────────────
	// Hapus aturan lama
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Aturan{}).Error; err != nil {
		log.Printf("⚠️ Gagal membersihkan tabel aturans: %v", err)
	} else {
		log.Println("🗑️ Tabel aturans berhasil dibersihkan untuk migrasi baru")
	}

	// Masukkan Aturan Gejala Inti Depresi & Kecemasan
	aturanData := []struct {
		KodePenyakit             string
		KodePertanyaan           string
		MinValue                 int64
		IsMandatory              int64
		TipeAturan               string
		BerlakuUntukSemuaTingkat bool
	}{
		// Gejala Inti Depresi
		{"P05", "G01", 2, 1, "GEJALA_INTI", false},
		{"P05", "G02", 2, 1, "GEJALA_INTI", false},
		{"P04", "G01", 2, 1, "GEJALA_INTI", false},
		{"P04", "G02", 1, 1, "GEJALA_INTI", false},
		{"P03", "G01", 1, 1, "GEJALA_INTI", false},
		{"P02", "G01", 0, 1, "GEJALA_INTI", false},
		{"P01", "G01", 0, 1, "GEJALA_INTI", false},

		// Gejala Inti Kecemasan
		{"K04", "D01", 2, 1, "GEJALA_INTI", false},
		{"K04", "D02", 2, 1, "GEJALA_INTI", false},
		{"K03", "D01", 1, 1, "GEJALA_INTI", false},
		{"K03", "D02", 1, 1, "GEJALA_INTI", false},
		{"K02", "D01", 0, 1, "GEJALA_INTI", false},
		{"K01", "D01", 0, 1, "GEJALA_INTI", false},

		// Red Flag (lintas semua tingkat)
		{"ALL", "G09", 1, 1, "RED_FLAG", true},
	}

	for _, a := range aturanData {
		newAturan := models.Aturan{
			KodePenyakit:             a.KodePenyakit,
			KodePertanyaan:           a.KodePertanyaan,
			MinValue:                 a.MinValue,
			IsMandatory:              a.IsMandatory,
			TipeAturan:               a.TipeAturan,
			BerlakuUntukSemuaTingkat: a.BerlakuUntukSemuaTingkat,
		}
		if err := db.Create(&newAturan).Error; err != nil {
			log.Printf("❌ Gagal membuat aturan (%s - %s): %v", a.KodePenyakit, a.KodePertanyaan, err)
		}
	}
	log.Println("🌱 Migrasi ulang aturan (aturans) berhasil dijalankan!")
}
