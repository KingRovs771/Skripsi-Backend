package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UpdateAdminInput struct {
	NamaLengkap string `json:"nama_lengkap"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Alamat      string `json:"alamat"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	RoleUID     string `json:"role_id"`
}
type DashboardSummary struct {
	// Health Status
	StatusSystem struct {
		ApiServer string `json:"api_server"`
		Database  string `json:"database"`
	} `json:"status_system"`

	// Main Stats
	Stats struct {
		TotalSiswa     int64 `json:"total_siswa"`
		TesSelesai     int64 `json:"tes_selesai"`
		ButuhPerhatian int64 `json:"butuh_perhatian"`
		SesiAktif      int64 `json:"sesi_aktif"`
	} `json:"stats"`

	// Chart Data
	Grafik []struct {
		Kategori string `json:"kategori"`
		Jumlah   int64  `json:"jumlah"`
	} `json:"grafik"`

	// Monitoring Data
	TrenMingguan []struct {
		Hari  string `json:"hari"`
		Total int64  `json:"total"`
	} `json:"tren_mingguan"`

	InstrumenStatus []struct {
		Nama   string `json:"nama"`
		Total  int64  `json:"total"`
		Status string `json:"status"`
	} `json:"instrumen_status"`
}

func GetFullDashboardData(c *gin.Context) {
	var data DashboardSummary

	sqlDB, err := database.DB.DB()
	data.StatusSystem.ApiServer = "ONLINE"
	if err != nil || sqlDB.Ping() != nil {
		data.StatusSystem.Database = "DISCONNECTED"
	} else {
		data.StatusSystem.Database = "CONNECTED"
	}

	// 1. Statistik Utama
	database.DB.Model(&models.Students{}).Count(&data.Stats.TotalSiswa)
	database.DB.Table("test_sessions").Where("status = ?", "SELESAI").Count(&data.Stats.TesSelesai)
	database.DB.Table("test_sessions").Where("status = ?", "BERJALAN").Count(&data.Stats.SesiAktif)
	database.DB.Table("hasil_diagnoses").Where("skor_total > ?", 15).Count(&data.Stats.ButuhPerhatian)

	// 2. Grafik Sebaran Penyakit — JOIN ke penyakits agar nama tampil, bukan kode
	database.DB.Table("hasil_diagnoses hd").
		Select("COALESCE(p.nama_penyakit, hd.final_depresi_penyakit) as kategori, count(*) as jumlah").
		Joins("LEFT JOIN penyakits p ON hd.final_depresi_penyakit = p.kode_penyakit").
		Group("hd.final_depresi_penyakit, p.nama_penyakit").
		Scan(&data.Grafik)

	// 3. Tren Mingguan
	database.DB.Raw(`
		SELECT TO_CHAR(created_at, 'Dy') as hari, count(*) as total 
		FROM test_sessions 
		WHERE created_at >= NOW() - INTERVAL '7 days' 
		GROUP BY hari, DATE_TRUNC('day', created_at)
		ORDER BY DATE_TRUNC('day', created_at)
	`).Scan(&data.TrenMingguan)

	// 4. Status Instrumen
	data.InstrumenStatus = []struct {
		Nama   string `json:"nama"`
		Total  int64  `json:"total"`
		Status string `json:"status"`
	}{
		{Nama: "PHQ-9 (Depresi)", Total: data.Stats.TesSelesai, Status: "Active"},
		{Nama: "GAD-7 (Kecemasan)", Total: data.Stats.TesSelesai, Status: "Active"},
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data":   data,
	})
}

func GetAllAdministrator(c *gin.Context) {
	var admins []models.Administrator

	if admins == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Data Not Found",
		})
		return
	}
	if err := database.DB.Find(&admins).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Tidak Dapat Mendapatkan Data Administrator",
			"Error":   err.Error(),
		})
		return
	}

	for i := range admins {
		admins[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Administrator",
		"Data":    admins,
	})
}

func CreateAdministrator(c *gin.Context) {
	var admin models.Administrator

	if err := c.ShouldBindJSON(&admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"Message": "HTTP Bad Request",
		})
		return
	}

	if err := admin.BeforeSaveAdministrator(database.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error":   err.Error(),
			"Message": "Status Internal Server Error",
		})
		return
	}
	saveAdmin, err := admin.SaveAdministrator()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error":   err.Error(),
			"Message": "Status Internal Server Error",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"Status":  "200 - ",
		"Message": "Administrator Berhasil Di Buat, Silakan check pada Halaman Utama Administrator",
		"Data":    saveAdmin,
	})
}

func GetAdministratorByUID(c *gin.Context) {
	var admin models.Administrator
	uid := c.Param("uid")

	if err := database.DB.Where("admin_uid = ?", uid).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"Message": "Administrator Not Found",
				"Error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"Message": "Database Error",
			"Error":   err.Error(),
		})
	}

	admin.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Administrator",
		"Data":    admin,
	})
}

func DeleteAdministrator(c *gin.Context) {
	uid := c.Param("uid")

	result := database.DB.Where("admin_uid = ?", uid).Delete(&models.Administrator{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Failed to Delete Administrator",
			"Error":   result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Not Found",
			"Message": "Administrator Not Found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Administrator Berhasil Di Hapus",
		"Data":    result,
	})

}

func UpdateAdministrator(c *gin.Context) {
	uid := c.Param("uid")

	var admin models.Administrator

	if err := database.DB.Where("admin_uid = ?", uid).First(&admin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Administrator Not Found",
		})
		return
	}

	var inputAdmin UpdateAdminInput
	if err := c.ShouldBindBodyWithJSON(&inputAdmin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Data",
			"Error":   err.Error(),
		})
		return
	}

	if inputAdmin.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(inputAdmin.Password), 12)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Status":  "Error",
				"Message": "Failed to hash password",
			})
			return
		}
		inputAdmin.Password = string(hashedPassword)
	}

	if err := database.DB.Model(&admin).Updates(inputAdmin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Failed to update administrator",
			"Error":   err.Error(),
		})
		return
	}

	admin.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Administrator updated successfully",
		"Data":    admin,
	})

}
