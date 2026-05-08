package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

// 1. Format Request dari Next.js (Saat klik "Mulai Tes")
type StartTestRequest struct {
	UserUID string `json:"user_uid" binding:"required"`
}

// 2. Format Response Soal untuk Frontend
type SoalResponse struct {
	KodePertanyaan string `json:"kode_pertanyaan"`
	Pertanyaan     string `json:"pertanyaan"`
}

type JawabanInput struct {
	KodePertanyaan string `json:"kode_pertanyaan"`
	Nilai          int64  `json:"nilai"`
}
type SubmitTestRequest struct {
	SessionID   string `json:"session_id" binding:"required"`
	UserUID     string `json:"user_uid" binding:"required"`
	CeritaSiswa string `json:"cerita_siswa"` // Bisa kosong jika siswa tidak mau cerita
	Jawaban     []struct {
		KodePertanyaan string `json:"kode_pertanyaan"`
		Nilai          int64  `json:"nilai"`
	} `json:"jawaban"`
}

// --- B. FORMAT DATA KE/DARI FASTAPI (PYTHON) ---
type AIPayload struct {
	SkorPHQ9 int64 `json:"skor_phq9"`
	SkorGAD7 int64 `json:"skor_gad7"`
}

type AIResponse struct {
	Depresi struct {
		Kode       string  `json:"kode"`
		Confidence float64 `json:"confidence"`
	} `json:"depresi"`
	Kecemasan struct {
		Kode       string  `json:"kode"`
		Confidence float64 `json:"confidence"`
	} `json:"kecemasan"`
}

func StartTest(c *gin.Context) {
	var req StartTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User UID tidak valid"})
		return
	}

	sessionID := uuid.New().String()
	sesiBaru := models.TestSession{
		TestSessionId:  sessionID,
		UserUID:        req.UserUID,
		TotalScorephq9: 0,
		TotalScoregad7: 0,
		Status:         "BERJALAN",
		CreatedAt:      time.Now(),
	}

	// Ganti config.DB dengan variabel global database Anda
	if err := database.DB.Create(&sesiBaru).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat sesi tes"})
		return
	}

	var daftarSoal []models.Pertanyaan
	if err := database.DB.Find(&daftarSoal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil bank soal"})
		return
	}

	var soalPHQ9, soalGAD7 []SoalResponse
	for _, soal := range daftarSoal {
		format := SoalResponse{KodePertanyaan: soal.KodePertanyaan, Pertanyaan: soal.Pertanyaan}
		if soal.KategoriPertanyaan == "PHQ9" {
			soalPHQ9 = append(soalPHQ9, format)
		} else {
			soalGAD7 = append(soalGAD7, format)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Sesi dimulai",
		"session_id": sessionID,
		"data_soal":  gin.H{"phq9": soalPHQ9, "gad7": soalGAD7},
	})
}

func SubmitTest(c *gin.Context) {
	var req SubmitTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	var sesi models.TestSession
	if err := database.DB.Where("test_session_id = ?", req.SessionID).First(&sesi).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sesi tidak ditemukan"})
		return
	}

	jawabanMap := make(map[string]int64)
	var totalPHQ9, totalGAD7 int64 = 0, 0

	for _, jwb := range req.Jawaban {
		jawabanMap[jwb.KodePertanyaan] = jwb.Nilai
		if jwb.KodePertanyaan[0] == 'G' {
			totalPHQ9 += jwb.Nilai
		} else if jwb.KodePertanyaan[0] == 'D' { // Sudah menggunakan huruf D sesuai database
			totalGAD7 += jwb.Nilai
		}

		database.DB.Create(&models.TestAnswer{
			TestAnswerId:   uuid.New().String(),
			TestSessionUID: req.SessionID,
			KodePertanyaan: jwb.KodePertanyaan,
			NilaiJawaban:   jwb.Nilai,
		})
	}

	database.DB.Model(&sesi).Updates(map[string]interface{}{
		"total_scorephq9": totalPHQ9,
		"total_scoregad7": totalGAD7,
		"status":          "SELESAI",
	})

	if req.CeritaSiswa != "" {
		database.DB.Create(&models.StudentFeedback{
			TestSessionUID: req.SessionID,
			CeritaSiswa:    req.CeritaSiswa,
			CreatedAt:      time.Now(),
		})
	}

	// =========================================================================
	// INTEGRASI FASTAPI (PYTHON NEURAL NETWORK)
	// =========================================================================
	aiPayload := AIPayload{SkorPHQ9: totalPHQ9, SkorGAD7: totalGAD7}
	jsonData, _ := json.Marshal(aiPayload)

	// Melakukan HTTP POST request ke server FastAPI
	resp, err := http.Post("http://localhost:8000/predict", "application/json", bytes.NewBuffer(jsonData))

	var aiResult AIResponse
	if err != nil {
		// Fallback/Skenario darurat jika server Python mati
		fmt.Println("Peringatan: Gagal terhubung ke FastAPI:", err)
		aiResult.Depresi.Kode = "P01"
		aiResult.Kecemasan.Kode = "K01"
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &aiResult); err != nil {
			fmt.Println("Peringatan: Gagal membaca balasan JSON dari AI:", err)
			aiResult.Depresi.Kode = "P01"
			aiResult.Kecemasan.Kode = "K01"
		}
	}

	// Memasukkan hasil balasan API ke dalam variabel existing
	aiDepresiKode := aiResult.Depresi.Kode
	aiCemasKode := aiResult.Kecemasan.Kode
	aiDepresiConf := aiResult.Depresi.Confidence
	aiCemasConf := aiResult.Kecemasan.Confidence
	// =========================================================================

	// Backward Chaining
	finalDepresi, statusDepresi := JalankanBackwardChaining(aiDepresiKode, jawabanMap)
	finalCemas, statusCemas := JalankanBackwardChaining(aiCemasKode, jawabanMap)

	if jawabanMap["G09"] >= 1 {
		statusDepresi = "URGENT_INTERVENTION"
		if finalDepresi == "P01" || finalDepresi == "P02" || finalDepresi == "P03" {
			finalDepresi = "P04"
		}
	}

	database.DB.Create(&models.HasilDiagnosis{
		ResultUID:             uuid.New().String(),
		SessionTestUID:        req.SessionID,
		NNDepresiPrediksi:     aiDepresiKode,
		NNDepresiConfidence:   aiDepresiConf,
		FinalDepresiPenyakit:  finalDepresi,
		StatusValidasiDepresi: statusDepresi,
		NNCemasPrediksi:       aiCemasKode,
		NNCemasConfidence:     aiCemasConf,
		FinalCemasPenyakit:    finalCemas,
		StatusValidasiCemas:   statusCemas,
		CreatedAt:             time.Now(),
	})

	// Mengirimkan respons secara langsung tanpa wrapper tambahan sesuai instruksi sebelumnya
	c.JSON(http.StatusOK, gin.H{
		"message":    "Tes berhasil",
		"session_id": req.SessionID,
	})
}
func JalankanBackwardChaining(tebakanAI string, jawabanSiswa map[string]int64) (string, string) {
	kodeSekarang := tebakanAI
	statusValidasi := "CONFIRMED"

	for kodeSekarang != "" {
		var penyakit models.Penyakit
		// Akses DB global
		err := database.DB.Preload("DaftarAturan").Where("kode_penyakit = ?", kodeSekarang).First(&penyakit).Error
		if err != nil {
			break
		}

		syaratTerpenuhi := true
		for _, aturan := range penyakit.DaftarAturan {
			if aturan.IsMandatory == 1 && jawabanSiswa[aturan.KodePertanyaan] < aturan.MinValue {
				syaratTerpenuhi = false
				break
			}
		}

		if syaratTerpenuhi {
			return kodeSekarang, statusValidasi
		}

		if penyakit.KodeTurunan != "" {
			kodeSekarang = penyakit.KodeTurunan
			statusValidasi = "ADJUSTED"
		} else {
			return kodeSekarang, "ADJUSTED"
		}
	}
	return kodeSekarang, statusValidasi
}
