package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Pakar Users
func CreatePakar(c *gin.Context) {
	var input struct {
		RoleUID        string `json:"role_uid" binding:"required"`
		NomorSIP       string `json:"nomor_sip" validate:"required"`
		NamaLengkap    string `json:"nama_lengkap" binding:"required"`
		JenisSpesialis string `json:"jenis_spesialis" binding:"required"`
		Phone          string `json:"phone" validate:"max=20"`
		Email          string `json:"email" binding:"required,email"`
		Alamat         string `json:"alamat"`
		Password       string `json:"password" binding:"required,min=6"`
		PhotoFile      []byte `json:"photo_file"` // Gambar dalam bentuk byte array
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pakar := models.Pakar{
		RoleUID:        input.RoleUID,
		NomorSIP:       input.NomorSIP,
		NamaLengkap:    input.NamaLengkap,
		JenisSpesialis: input.JenisSpesialis,
		Phone:          input.Phone,
		Email:          input.Email,
		Alamat:         input.Alamat,
		Password:       input.Password,
		PhotoFile:      input.PhotoFile,
	}

	result, err := pakar.SaveUsersPakar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Data pakar berhasil ditambahkan",
		"data":    result,
	})
}

func GetAllPakar(c *gin.Context) {
	pakarList, err := models.GetAllPakar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	baseURL := "http://" + c.Request.Host

	for i := range pakarList {
		if len(pakarList[i].PhotoFile) > 0 {
			pakarList[i].PhotoURL = fmt.Sprintf("%s/api/photo/getPhotoPakar/%s", baseURL, pakarList[i].PakarUID)
		} else {
			pakarList[i].PhotoURL = ""
		}

	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data pakar berhasil diambil",
		"data":    pakarList,
	})
}

func GetPakarPhoto(c *gin.Context) {
	uid := c.Param("uid")
	var pakar models.Pakar

	if err := database.DB.Select("photo_file").Where("pakar_uid = ?", uid).First(&pakar).Error; err != nil {
		c.Data(http.StatusNotFound, "text/plain", []byte("Foto tidak ditemukan"))
		return
	}

	if len(pakar.PhotoFile) == 0 {
		c.Data(http.StatusNotFound, "text/plain", []byte("Data foto kosong"))
		return
	}

	contentType := http.DetectContentType(pakar.PhotoFile)

	c.Data(http.StatusOK, contentType, pakar.PhotoFile)
}

func GetPakarByUID(c *gin.Context) {
	uid := c.Param("uid")

	pakar, err := models.GetPakarByUID(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pakar tidak ditemukan"})
		return
	}

	// Jangan kirim PhotoFile ke client
	pakar.PhotoFile = nil

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data pakar berhasil diambil",
		"data":    pakar,
	})
}

func UpdatePakar(c *gin.Context) {
	uid := c.Param("uid")

	var input struct {
		RoleUID        *string `json:"role_uid,omitempty"`
		NomorSIP       *string `json:"nomor_sip,omitempty" binding:"omitempty"`
		NamaLengkap    *string `json:"nama_lengkap,omitempty" binding:"omitempty"`
		JenisSpesialis *string `json:"jenis_spesialis,omitempty"`
		Phone          *string `json:"phone,omitempty" binding:"omitempty,max=20"`
		Email          *string `json:"email,omitempty" binding:"omitempty,email"`
		Alamat         *string `json:"alamat,omitempty"`
		PhotoFile      *[]byte `json:"photo_file,omitempty"` // Optional update photo
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pakar models.Pakar
	if err := database.DB.Where("pakar_uid = ?", uid).First(&pakar).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pakar tidak ditemukan"})
		return
	}

	// Update fields
	if input.RoleUID != nil {
		pakar.RoleUID = *input.RoleUID
	}
	if input.NomorSIP != nil {
		pakar.NomorSIP = *input.NomorSIP
	}
	if input.NamaLengkap != nil {
		pakar.NamaLengkap = *input.NamaLengkap
	}
	if input.JenisSpesialis != nil {
		pakar.JenisSpesialis = *input.JenisSpesialis
	}
	if input.Phone != nil {
		pakar.Phone = *input.Phone
	}
	if input.Email != nil {
		pakar.Email = *input.Email
	}
	if input.Alamat != nil {
		pakar.Alamat = *input.Alamat
	}
	if input.PhotoFile != nil {
		pakar.PhotoFile = *input.PhotoFile
	}

	if err := pakar.UpdatePakar(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Jangan kirim PhotoFile ke client
	pakar.PhotoFile = nil

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data pakar berhasil diperbarui",
		"data":    pakar,
	})
}

func DeletePakar(c *gin.Context) {
	uid := c.Param("uid")

	err := models.DeletePakar(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data pakar berhasil dihapus",
	})
}

// Akun Siswa
func CreateStudents(c *gin.Context) {
	var inputStudents struct {
		RoleUID           string `json:"role_uid" binding:"required"`
		NISN              string `json:"nisn" binding:"required"`
		NamaLengkap       string `json:"nama_lengkap" binding:"required"`
		NPSN              string `json:"npsn" binding:"required"`
		JenjangPendidikan string `json:"jenjang_pendidikan" binding:"required"`
		Kelas             string `json:"kelas" binding:"required"`
		Email             string `json:"email" binding:"required"`
		Password          string `json:"password" binding:"required"`
		NoHp              string `json:"no_hp" binding:"required"`
		Alamat            string `json:"alamat" binding:"required"`
	}

	if err := c.ShouldBindJSON(&inputStudents); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  http.StatusBadRequest,
			"Message": "Invalid Input, Silakan Ulangi Lagi",
			"Error":   err,
		})
		return
	}

	st := models.Students{
		RoleUID:     inputStudents.RoleUID,
		NISN:        inputStudents.NISN,
		NamaLengkap: inputStudents.NamaLengkap,
		NoHp:        inputStudents.NoHp,
		Alamat:      inputStudents.Alamat,
		NPSN:        inputStudents.NPSN,
		Kelas:       inputStudents.Kelas,
		Email:       inputStudents.Email,
		Password:    inputStudents.Password,
	}
	savedStudents, err := st.Save()
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

func GetAllStudents(c *gin.Context) {
	StudentsList, err := models.GetAllStudents()
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
		"Message": "Data Siswa berhasil Didapatkan",
		"Data":    StudentsList,
	})
}

func GetStudentById(c *gin.Context) {
	uid := c.Param("uid")

	student, err := models.GetStudentsById(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pakar tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data pakar berhasil diambil",
		"data":    student,
	})
}

func DeleteStudent(c *gin.Context) {
	uid := c.Param("uid")

	err := models.DeleteStudents(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  http.StatusNotFound,
			"Message": "Akun Siswa tidak ditemukan",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  http.StatusOK,
		"Message": "Akun Siswa Berhasil Dihapus",
		"Data":    err,
	})
}

func UpdateStudents(c *gin.Context) {
	uid := c.Param("uid")

	var input struct {
		RoleUID           *string `json:"role_uid" binding:"required"`
		NISN              *string `json:"nisn,omitempty" binding:"omitempty,max=20"`
		NamaLengkap       *string `json:"nama_lengkap,omitempty" binding:"omitempty,max=90"`
		JenjangPendidikan *string `json:"jenjang_pendidikan,omitempty"`
		Kelas             *string `json:"kelas,omitempty"`
		NoHp              *string `json:"no_hp,omitempty" binding:"omitempty,max=20"`
		Alamat            *string `json:"alamat,omitempty"`
		Email             *string `json:"email,omitempty" binding:"omitempty,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ambil data siswa berdasarkan UID
	var student models.Students
	db := database.DB
	if err := db.Where("students_uid = ?", uid).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Siswa tidak ditemukan"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Update field-field yang diinputkan
	if input.NISN != nil {
		student.NISN = *input.NISN
	}
	if input.NamaLengkap != nil {
		student.NamaLengkap = *input.NamaLengkap
	}
	if input.Kelas != nil {
		student.Kelas = *input.Kelas
	}
	if input.NoHp != nil {
		student.NoHp = *input.NoHp
	}
	if input.Alamat != nil {
		student.Alamat = *input.Alamat
	}
	if input.Email != nil {
		student.Email = *input.Email
	}

	// Jalankan update ke database
	if err := db.Save(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data siswa"})
		return
	}

	// Kirim response sukses
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data siswa berhasil diperbarui",
		"data":    student,
	})
}

// Akun Guru
func CreateTeachers(c *gin.Context) {
	var input struct {
		RoleUID     string `json:"role_uid" binding:"required"`
		NIP         string `json:"nip" binding:"required,max=20"`
		NamaLengkap string `json:"nama_lengkap" binding:"required,max=90"`
		NPSN        string `json:"npsn" binding:"max=20"`
		Phone       string `json:"phone" binding:"max=20"`
		Alamat      string `json:"alamat"`
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=6"`
		PhotoFile   []byte `json:"photo_file"` // Gambar dalam bentuk byte array
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacher := models.Teachers{
		RoleUID:     input.RoleUID,
		NIP:         input.NIP,
		NamaLengkap: input.NamaLengkap,
		NPSN:        input.NPSN,
		Phone:       input.Phone,
		Alamat:      input.Alamat,
		Email:       input.Email,
		Password:    input.Password,
		PhotoFile:   input.PhotoFile,
	}

	result, err := teacher.SaveTeachers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Data guru berhasil ditambahkan",
		"data":    result,
	})
}

// GetAllTeachers
func GetAllTeachers(c *gin.Context) {
	teachersList, err := models.GetAllTeachers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Hilangkan PhotoFile dari response JSON
	for i := range teachersList {
		teachersList[i].PhotoFile = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data guru berhasil diambil",
		"data":    teachersList,
	})
}

// GetTeachersByUID
func GetTeachersByUID(c *gin.Context) {
	uid := c.Param("uid")

	teacher, err := models.GetTeachersByUID(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guru tidak ditemukan"})
		return
	}

	// Jangan kirim PhotoFile ke client
	teacher.PhotoFile = nil

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data guru berhasil diambil",
		"data":    teacher,
	})
}

// UpdateTeachers
func UpdateTeachers(c *gin.Context) {
	uid := c.Param("uid")

	var input struct {
		RoleUID     *string `json:"role_uid,omitempty"`
		NIP         *string `json:"nip,omitempty" binding:"omitempty,max=20"`
		NamaLengkap *string `json:"nama_lengkap,omitempty" binding:"omitempty,max=90"`
		NPSN        *string `json:"npsn,omitempty" binding:"omitempty,max=20"`
		Phone       *string `json:"phone,omitempty" binding:"omitempty,max=20"`
		Alamat      *string `json:"alamat,omitempty"`
		Email       *string `json:"email,omitempty" binding:"omitempty,email"`
		PhotoFile   *[]byte `json:"photo_file,omitempty"` // Optional update photo
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var teacher models.Teachers
	db := database.DB
	if err := db.Where("teachers_uid = ?", uid).First(&teacher).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guru tidak ditemukan"})
		return
	}

	// Update fields
	if input.RoleUID != nil {
		teacher.RoleUID = *input.RoleUID
	}
	if input.NIP != nil {
		teacher.NIP = *input.NIP
	}
	if input.NamaLengkap != nil {
		teacher.NamaLengkap = *input.NamaLengkap
	}
	if input.NPSN != nil {
		teacher.NPSN = *input.NPSN
	}
	if input.Phone != nil {
		teacher.Phone = *input.Phone
	}
	if input.Alamat != nil {
		teacher.Alamat = *input.Alamat
	}
	if input.Email != nil {
		teacher.Email = *input.Email
	}
	if input.PhotoFile != nil {
		teacher.PhotoFile = *input.PhotoFile
	}

	if err := teacher.UpdateTeachers(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Jangan kirim PhotoFile ke client
	teacher.PhotoFile = nil

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data guru berhasil diperbarui",
		"data":    teacher,
	})
}

// DeleteTeachers
func DeleteTeachers(c *gin.Context) {
	uid := c.Param("uid")

	err := models.DeleteTeachers(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Data guru berhasil dihapus",
	})
}
