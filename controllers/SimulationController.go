package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DraftAturanInput struct {
	KodePenyakit   string `json:"kode_penyakit" binding:"required"`
	KodePertanyaan string `json:"kode_pertanyaan" binding:"required"`
	MinValue       int64  `json:"min_value"`
	IsMandatory    int64  `json:"is_mandatory"`
}

type SimulationRequest struct {
	SampleSize int64              `json:"sample_size" binding:"required,min=1,max=200"`
	DraftRules []DraftAturanInput `json:"draft_rules"`
}

type SimulatedSessionDetail struct {
	SessionID        string    `json:"session_id"`
	Tanggal          time.Time `json:"tanggal"`
	OriginalDepresi  string    `json:"original_depresi"`
	SimulatedDepresi string    `json:"simulated_depresi"`
	OriginalCemas    string    `json:"original_cemas"`
	SimulatedCemas    string    `json:"simulated_cemas"`
	IsDifferent      bool      `json:"is_different"`
}

// SimulateRules runs the backward chaining engine on historical sessions using draft rules
// POST /api/pakar/aturans/simulasi
func SimulateRules(c *gin.Context) {
	var req SimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Format payload tidak valid",
		})
		return
	}

	// 1. Ambil sampel sesi tes historis yang sudah selesai
	var sessions []models.TestSession
	err := database.DB.Where("status = 'SELESAI'").
		Order("created_at DESC").
		Limit(int(req.SampleSize)).
		Find(&sessions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil sampel sesi tes historis",
		})
		return
	}

	if len(sessions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"Status": "Success",
			"Data": gin.H{
				"total_tested": 0,
				"total_same":   0,
				"total_diff":   0,
				"details":      []SimulatedSessionDetail{},
			},
		})
		return
	}

	// Group draft rules by disease code for quick override lookup
	draftMap := make(map[string][]models.Aturan)
	for _, rule := range req.DraftRules {
		draftMap[rule.KodePenyakit] = append(draftMap[rule.KodePenyakit], models.Aturan{
			KodePenyakit:   rule.KodePenyakit,
			KodePertanyaan: rule.KodePertanyaan,
			MinValue:       rule.MinValue,
			IsMandatory:    rule.IsMandatory,
		})
	}

	var details []SimulatedSessionDetail
	var totalSame, totalDiff int

	for _, session := range sessions {
		// Ambil hasil diagnosis asli dari database
		var originalResult models.HasilDiagnosis
		if err := database.DB.Where("session_test_uid = ?", session.TestSessionId).First(&originalResult).Error; err != nil {
			continue // Lewati jika tidak ada hasil diagnosis
		}

		// Ambil semua jawaban siswa untuk sesi ini
		var answers []models.TestAnswer
		if err := database.DB.Where("test_session_uid = ?", session.TestSessionId).Find(&answers).Error; err != nil {
			continue
		}

		// Buat map jawaban
		jawabanMap := make(map[string]int64)
		for _, ans := range answers {
			jawabanMap[ans.KodePertanyaan] = ans.NilaiJawaban
		}

		// Jalankan Backward Chaining Simulasi (Depresi & Cemas)
		simDepresi := JalankanBackwardChainingSimulasi(originalResult.NNDepresiPrediksi, jawabanMap, draftMap)
		simCemas := JalankanBackwardChainingSimulasi(originalResult.NNCemasPrediksi, jawabanMap, draftMap)

		// Terapkan G09 Urgent Intervention rule konsisten dengan logic asli
		if jawabanMap["G09"] >= 1 {
			if simDepresi == "P01" || simDepresi == "P02" || simDepresi == "P03" {
				simDepresi = "P04"
			}
		}

		isDiff := simDepresi != originalResult.FinalDepresiPenyakit || simCemas != originalResult.FinalCemasPenyakit
		if isDiff {
			totalDiff++
		} else {
			totalSame++
		}

		details = append(details, SimulatedSessionDetail{
			SessionID:        session.TestSessionId,
			Tanggal:          session.CreatedAt,
			OriginalDepresi:  originalResult.FinalDepresiPenyakit,
			SimulatedDepresi: simDepresi,
			OriginalCemas:    originalResult.FinalCemasPenyakit,
			SimulatedCemas:   simCemas,
			IsDifferent:      isDiff,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data": gin.H{
			"total_tested": len(details),
			"total_same":   totalSame,
			"total_diff":   totalDiff,
			"details":      details,
		},
	})
}

// JalankanBackwardChainingSimulasi melakukan inferensi backward chaining dengan mensubstitusi
// aturan dari database dengan aturan draf dari pakar jika ada.
func JalankanBackwardChainingSimulasi(tebakanAI string, jawabanSiswa map[string]int64, draftMap map[string][]models.Aturan) string {
	kodeSekarang := tebakanAI

	for kodeSekarang != "" {
		var rules []models.Aturan
		var penyakit models.Penyakit

		// 1. Ambil info penyakit
		err := database.DB.Where("kode_penyakit = ?", kodeSekarang).First(&penyakit).Error
		if err != nil {
			break
		}

		// 2. Ambil aturan: gunakan draft jika ada override, jika tidak ambil dari database
		if overrideRules, ok := draftMap[kodeSekarang]; ok {
			rules = overrideRules
		} else {
			database.DB.Where("kode_penyakit = ?", kodeSekarang).Find(&rules)
		}

		syaratTerpenuhi := true
		for _, aturan := range rules {
			if aturan.IsMandatory == 1 && jawabanSiswa[aturan.KodePertanyaan] < aturan.MinValue {
				syaratTerpenuhi = false
				break
			}
		}

		if syaratTerpenuhi {
			return kodeSekarang
		}

		if penyakit.KodeTurunan != "" {
			kodeSekarang = penyakit.KodeTurunan
		} else {
			return kodeSekarang
		}
	}
	return kodeSekarang
}
