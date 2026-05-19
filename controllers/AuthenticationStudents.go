package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterStudents(c *gin.Context) {
	var registerUser struct {
		RoleUID           string `json:"role_uid"`
		NISN              string `json:"nisn"`
		NamaLengkap       string `json:"nama_lengkap"`
		NoHp              string `json:"no_hp"`
		Alamat            string `json:"alamat"`
		NPSN              string `json:"npsn"`
		JenjangPendidikan string `json:"jenjang_pendidikan"`
		Kelas             string `json:"kelas"`
		Email             string `json:"email" validate:"required,email"`
		Password          string `json:"password" validate:"required, min=8"`
	}

	if err := c.BindJSON(&registerUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Error",
			"Error":   err.Error(),
		})
		return
	}

	// Cek duplikat NISN
	var nisnCount int64
	database.DB.Model(&models.Students{}).Where("nisn = ?", registerUser.NISN).Count(&nisnCount)
	if nisnCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"Status":  "Error",
			"Message": "NISN sudah terdaftar, gunakan NISN yang berbeda atau lakukan login",
		})
		return
	}

	// Cek duplikat Email
	var emailCount int64
	database.DB.Model(&models.Students{}).Where("email = ?", registerUser.Email).Count(&emailCount)
	if emailCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"Status":  "Error",
			"Message": "Email sudah terdaftar, gunakan email lain atau lakukan login",
		})
		return
	}

	registerStudents := models.Students{
		RoleUID:     registerUser.RoleUID,
		NISN:        registerUser.NISN,
		NamaLengkap: registerUser.NamaLengkap,
		NoHp:        registerUser.NoHp,
		Alamat:      registerUser.Alamat,
		NPSN:        registerUser.NPSN,
		Kelas:       registerUser.Kelas,
		Email:       registerUser.Email,
		Password:    registerUser.Password,
	}
	savedStudents, err := registerStudents.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal Error",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Register Berhasil, Silakan Lanjutkan Proses Login",
		"Data":    savedStudents,
	})

}

func LoginStudents(c *gin.Context) {
	var loginUser struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&loginUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Error",
			"Error":   err.Error(),
		})
		return
	}

	loginStudents, err := models.FindUserByEmail(loginUser.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"Status":  "Error",
				"Message": "User Not Found",
				"Error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal Error",
			"Error":   err.Error(),
		})
		return
	}
	err = loginStudents.ValidatePassword(loginUser.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Email atau Password Salah, Silakan Coba Lagi",
			"Error":   err.Error(),
		})
		return
	}
	jwt, err := utils.GenerateJWT(loginStudents.StudentsUID, loginStudents.Email, "Students")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal Error",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Login Success",
		"Token":   jwt,
		"Data": gin.H{
			"student_uid":  loginStudents.StudentsUID,
			"nisn":         loginStudents.NISN,
			"nama_lengkap": loginStudents.NamaLengkap,
			"email":        loginStudents.Email,
		},
	})
}

func GetProfileStudents(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Invalid Token",
			"Error":   err.Error(),
		})
		return
	}
	studentsUID := claims.ID
	var students models.Students
	if err := database.DB.Select(`students.students_id, students.students_uid, students.role_uid,
				students.nisn, students.nama_lengkap, students.npsn, students.kelas, students.no_hp,
				students.alamat, students.email, students.created_at, students.update_at,
				roles.role_name, sekolahs.nama_sekolah`).
		Joins("left join roles on roles.role_uid = students.role_uid").
		Joins("left join sekolahs on sekolahs.npsn::text = students.npsn").
		Where("students_uid = ?", studentsUID).First(&students).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Invalid Token",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Success",
		"Data":    students,
	})
}
