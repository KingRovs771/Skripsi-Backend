package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// --- STUDENT SIDE ---

// AskQuestion handles student asking a question
func AskQuestion(c *gin.Context) {
	var input struct {
		Pertanyaan string `json:"pertanyaan" binding:"required"`
		Tujuan     string `json:"tujuan" binding:"required"` // BK atau PAKAR
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "Error", "Message": "Input tidak valid"})
		return
	}

	// Ambil user_uid dari context (set by middleware)
	userUID, exists := c.Get("user_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"Status": "Error", "Message": "Tidak terautentikasi"})
		return
	}

	faq := models.Faqs{
		UserUID:       userUID.(string),
		Tujuan:        input.Tujuan,
		FaqPertanyaan: input.Pertanyaan,
		Status:        "Menunggu Balasan",
	}

	if err := database.DB.Create(&faq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "Error", "Message": "Gagal mengirim pertanyaan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Pertanyaan berhasil dikirim",
		"Data":    faq,
	})
}

// GetStudentQuestions returns questions asked by the logged in student
func GetStudentQuestions(c *gin.Context) {
	userUID, exists := c.Get("user_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"Status": "Error", "Message": "Tidak terautentikasi"})
		return
	}

	var faqs []models.Faqs
	database.DB.Where("user_uid = ?", userUID).Order("created_at desc").Find(&faqs)

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"data":   faqs,
	})
}

// --- PAKAR / ADMIN SIDE ---

// GetAllFaqs returns all questions for Pakar/Admin
func GetAllFaqs(c *gin.Context) {
	type FaqDisplay struct {
		FaqsUID       string    `json:"faqs_uid"`
		NISN          string    `json:"nisn"`
		NamaLengkap   string    `json:"nama_lengkap"`
		FaqPertanyaan string    `json:"faq_pertanyaan"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
	}

	var results []FaqDisplay

	// Join dengan tabel students untuk mendapatkan NISN dan Nama
	err := database.DB.Table("faqs").
		Select("faqs.faqs_uid, students.nisn, students.nama_lengkap, faqs.faq_pertanyaan, faqs.status, faqs.created_at").
		Joins("left join students on students.students_uid = faqs.user_uid").
		Order("faqs.created_at desc").
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "Error", "Message": "Gagal mengambil data FAQ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data":   results,
	})
}

// GetGuruBKFaqs returns questions for Guru BK based on NPSN
func GetGuruBKFaqs(c *gin.Context) {
	userUID, exists := c.Get("user_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"Status": "Error", "Message": "Tidak terautentikasi"})
		return
	}

	var teacher models.Teachers
	if err := database.DB.Where("n_ip = ?", userUID).First(&teacher).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "Error", "Message": "Gagal mengambil profil guru"})
		return
	}

	type FaqDisplay struct {
		FaqsUID       string    `json:"faqs_uid"`
		NISN          string    `json:"nisn"`
		NamaLengkap   string    `json:"nama_lengkap"`
		FaqPertanyaan string    `json:"faq_pertanyaan"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
	}

	var results []FaqDisplay

	// Ambil FAQ yang ditujukan ke BK dan berasal dari sekolah yang sama (NPSN)
	err := database.DB.Table("faqs").
		Select("faqs.faqs_uid, students.nisn, students.nama_lengkap, faqs.faq_pertanyaan, faqs.status, faqs.created_at").
		Joins("inner join students on students.students_uid = faqs.user_uid").
		Where("students.npsn = ? AND faqs.tujuan = ?", teacher.NPSN, "BK").
		Order("faqs.created_at desc").
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "Error", "Message": "Gagal mengambil data FAQ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data":   results,
	})
}

// ReplyFaq handles replying to a question
func ReplyFaq(c *gin.Context) {
	uid := c.Param("uid")
	var input struct {
		Jawaban string `json:"jawaban" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "Error", "Message": "Jawaban wajib diisi"})
		return
	}

	userUID, exists := c.Get("user_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"Status": "Error", "Message": "Tidak terautentikasi"})
		return
	}

	var faq models.Faqs
	if err := database.DB.Where("faqs_uid = ?", uid).First(&faq).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Status": "Error", "Message": "Pertanyaan tidak ditemukan"})
		return
	}

	faq.FaqJawaban = input.Jawaban
	faq.Status = "Terjawab"
	faq.RepliedBy = userUID.(string)
	faq.UpdateAt = time.Now()

	if err := database.DB.Save(&faq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "Error", "Message": "Gagal menyimpan jawaban"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Jawaban berhasil dikirim",
	})
}

// GetFaqByUID returns a single FAQ
func GetFaqByUID(c *gin.Context) {
	uid := c.Param("uid")

	type FaqDetail struct {
		FaqsUID       string    `json:"faqs_uid"`
		NISN          string    `json:"nisn"`
		NamaLengkap   string    `json:"nama_lengkap"`
		FaqPertanyaan string    `json:"faq_pertanyaan"`
		FaqJawaban    string    `json:"faq_jawaban"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
	}

	var result FaqDetail
	err := database.DB.Table("faqs").
		Select("faqs.faqs_uid, students.nisn, students.nama_lengkap, faqs.faq_pertanyaan, faqs.faq_jawaban, faqs.status, faqs.created_at").
		Joins("left join students on students.students_uid = faqs.user_uid").
		Where("faqs.faqs_uid = ?", uid).
		Scan(&result).Error

	if err != nil || result.FaqsUID == "" {
		c.JSON(http.StatusNotFound, gin.H{"Status": "Error", "Message": "Data tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data":   result,
	})
}

// DeleteFaq deletes a question
func DeleteFaq(c *gin.Context) {
	uid := c.Param("uid")

	if err := database.DB.Where("faqs_uid = ?", uid).Delete(&models.Faqs{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "Error", "Message": "Gagal menghapus data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Data berhasil dihapus",
	})
}
