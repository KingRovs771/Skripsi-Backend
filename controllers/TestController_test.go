package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestJalankanBackwardChaining(t *testing.T) {
	// 1. Load .env
	_ = godotenv.Load("../.env")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL tidak di-set, melewati integration test.")
		return
	}

	// 2. Koneksi ke Database secara aman (tanpa crash jika port tertutup)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Gagal terhubung ke database PostgreSQL (%s): %v. Melewati integration test.", dsn, err)
		return
	}
	originalDB := database.DB
	defer func() {
		database.DB = originalDB
	}()

	// 3. Mulai Transaksi agar tidak merusak data riil
	tx := db.Begin()
	defer tx.Rollback()
	database.DB = tx

	// 4. Bersihkan tabel uji
	tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Aturan{})
	tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Penyakit{})

	// 5. Setup data mock Penyakit (PHQ-9 & GAD-7)
	p5Min, p5Max := int64(20), int64(27)
	p4Min, p4Max := int64(15), int64(19)
	p3Min, p3Max := int64(10), int64(14)
	p2Min, p2Max := int64(5), int64(9)
	p1Min, p1Max := int64(0), int64(4)

	penyakitMock := []models.Penyakit{
		{KodePenyakit: "P05", NamaPenyakit: "Depresi Berat", MinSkor: &p5Min, MaxSkor: &p5Max, KodeTurunan: "P04"},
		{KodePenyakit: "P04", NamaPenyakit: "Depresi Sedang Berat", MinSkor: &p4Min, MaxSkor: &p4Max, KodeTurunan: "P03"},
		{KodePenyakit: "P03", NamaPenyakit: "Depresi Sedang", MinSkor: &p3Min, MaxSkor: &p3Max, KodeTurunan: "P02"},
		{KodePenyakit: "P02", NamaPenyakit: "Depresi Ringan", MinSkor: &p2Min, MaxSkor: &p2Max, KodeTurunan: "P01"},
		{KodePenyakit: "P01", NamaPenyakit: "Depresi Minimal / Normal", MinSkor: &p1Min, MaxSkor: &p1Max, KodeTurunan: ""},
	}
	for _, p := range penyakitMock {
		if err := tx.Create(&p).Error; err != nil {
			t.Fatalf("Gagal seeding penyakit mock: %v", err)
		}
	}

	// 6. Setup data mock Aturan (PHQ-9 & GAD-7 & Red Flag)
	aturanMock := []models.Aturan{
		// Gejala Inti: P05 wajib G01>=2 dan G02>=2
		{KodePenyakit: "P05", KodePertanyaan: "G01", MinValue: 2, IsMandatory: 1, TipeAturan: "GEJALA_INTI"},
		{KodePenyakit: "P05", KodePertanyaan: "G02", MinValue: 2, IsMandatory: 1, TipeAturan: "GEJALA_INTI"},

		// Gejala Inti: P04 wajib G01>=2 dan G02>=1
		{KodePenyakit: "P04", KodePertanyaan: "G01", MinValue: 2, IsMandatory: 1, TipeAturan: "GEJALA_INTI"},
		{KodePenyakit: "P04", KodePertanyaan: "G02", MinValue: 1, IsMandatory: 1, TipeAturan: "GEJALA_INTI"},

		// Gejala Inti: P03 wajib G01>=1
		{KodePenyakit: "P03", KodePertanyaan: "G01", MinValue: 1, IsMandatory: 1, TipeAturan: "GEJALA_INTI"},

		// Gejala Inti: P02 wajib G01>=0
		{KodePenyakit: "P02", KodePertanyaan: "G01", MinValue: 0, IsMandatory: 1, TipeAturan: "GEJALA_INTI"},

		// Gejala Inti: P01 wajib G01>=0
		{KodePenyakit: "P01", KodePertanyaan: "G01", MinValue: 0, IsMandatory: 1, TipeAturan: "GEJALA_INTI"},

		// Red Flag (Lintas Semua Tingkat)
		{KodePenyakit: "ALL", KodePertanyaan: "G09", MinValue: 1, IsMandatory: 1, TipeAturan: "RED_FLAG", BerlakuUntukSemuaTingkat: true},
	}
	for _, a := range aturanMock {
		if err := tx.Create(&a).Error; err != nil {
			t.Fatalf("Gagal seeding aturan mock: %v", err)
		}
	}

	// 7. UJI SKENARIO TEST CASES

	// (a) Lapisan 1 dan 2 sama-sama lolos di percobaan pertama tanpa backtrack
	// Tebakan AI P05, Skor total 24 (PHQ-9), G01=3, G02=3, G09=0
	t.Run("Skenario A - Lolos tanpa backtrack", func(t *testing.T) {
		jawaban := map[string]int64{
			"G01": 3,
			"G02": 3,
			"G03": 3,
			"G04": 3,
			"G05": 3,
			"G06": 3,
			"G07": 3,
			"G08": 3,
			"G09": 0,
		}
		result, status := JalankanBackwardChaining("P05", jawaban)
		if result != "P05" || status != "CONFIRMED" {
			t.Errorf("Diharapkan P05 CONFIRMED, didapat %s %s", result, status)
		}
	})

	// (b) Lapisan 1 gagal (skor diluar rentang) tapi Lapisan 2 lolos (gejala inti terpenuhi)
	// Tebakan AI P05, Skor total 18 (PHQ-9 - masuk rentang P04), G01=2, G02=2
	t.Run("Skenario B - Lapisan 1 gagal tapi Lapisan 2 lolos", func(t *testing.T) {
		jawaban := map[string]int64{
			"G01": 2, // Lolos gejala inti P05
			"G02": 2, // Lolos gejala inti P05
			"G03": 2,
			"G04": 2,
			"G05": 2,
			"G06": 2,
			"G07": 2,
			"G08": 2,
			"G09": 0, // total score = 16 (PHQ-9) -> Rentang P04
		}
		result, status := JalankanBackwardChaining("P05", jawaban)
		if result != "P04" || status != "ADJUSTED" {
			t.Errorf("Diharapkan P04 ADJUSTED karena skor 16 berada di rentang P04, didapat %s %s", result, status)
		}
	})

	// (c) Lapisan 1 lolos (skor masuk rentang P05) tapi Lapisan 2 gagal (salah satu gejala inti tidak terpenuhi)
	// Tebakan AI P05, Skor total 20 (PHQ-9), G01=3, G02=0 (gagal gejala inti P05), backtrack ke P04 (skor 20 di luar P04), backtrack ke P03, dst.
	t.Run("Skenario C - Lapisan 1 lolos tapi Lapisan 2 gagal", func(t *testing.T) {
		jawaban := map[string]int64{
			"G01": 3,
			"G02": 0, // Gagal Lapisan 2 untuk P05 (butuh G02>=2) dan P04 (butuh G02>=1)
			"G03": 3,
			"G04": 3,
			"G05": 3,
			"G06": 3,
			"G07": 3,
			"G08": 2,
			"G09": 0, // total score = 20 -> Masuk rentang P05 tapi gagal gejala inti.
		}
		// Backtrack dari P05 ke P04 (skor 20 diluar P04) -> ke P03 (skor 20 diluar P03) -> ke P02 (skor 20 diluar P02) -> ke P01
		result, status := JalankanBackwardChaining("P05", jawaban)
		if result != "P01" || status != "ADJUSTED" {
			t.Errorf("Diharapkan backtrack ke P01 ADJUSTED karena skor 20 diluar rentang tingkat di bawahnya, didapat %s %s", result, status)
		}
	})

	// (d) Backtrack berlapis hingga mencapai P01
	// Tebakan AI P05, Skor total 2 (PHQ-9 - masuk rentang P01), G01=0, G02=0, G09=0
	t.Run("Skenario D - Backtrack berlapis ke P01", func(t *testing.T) {
		jawaban := map[string]int64{
			"G01": 0,
			"G02": 0,
			"G03": 1,
			"G04": 1,
			"G09": 0, // total score = 2 (Rentang P01)
		}
		result, status := JalankanBackwardChaining("P05", jawaban)
		if result != "P01" || status != "ADJUSTED" {
			t.Errorf("Diharapkan P01 ADJUSTED, didapat %s %s", result, status)
		}
	})

	// (e) Red Flag override terjadi meski hipotesis akhirnya P01
	// Tebakan AI P05, Skor total 1 (PHQ-9 - masuk rentang P01), G09=1 (Red Flag!)
	t.Run("Skenario E - Red Flag override ke P05", func(t *testing.T) {
		jawaban := map[string]int64{
			"G01": 0,
			"G02": 0,
			"G09": 1, // Red Flag pemicu ide bunuh diri
		}
		result, status := JalankanBackwardChaining("P05", jawaban)
		if result != "P05" || status != "URGENT_INTERVENTION" {
			t.Errorf("Diharapkan P05 URGENT_INTERVENTION karena Red Flag terpicu, didapat %s %s", result, status)
		}
	})
}
