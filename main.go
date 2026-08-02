package main

import (
	"Skripsi-Backend/controllers"
	"Skripsi-Backend/database"
	"Skripsi-Backend/middleware"
	"Skripsi-Backend/models"
	"Skripsi-Backend/seeder"
	"Skripsi-Backend/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Di Railway, env vars di-inject langsung oleh platform tanpa file .env.
	// Error diabaikan agar app tetap berjalan di Railway maupun lokal.
	// Jika file .env ada (development lokal), variabelnya tetap terbaca normal.
	_ = godotenv.Load()

	database.Connect()
	database.ConnectRedis()

	// Start background retention scheduler (Fitur 1B)
	utils.StartRetentionScheduler()
	_ = database.DB.AutoMigrate(
		&models.Penyakit{},
		&models.Administrator{},
		&models.Article{},
		&models.Aturan{},
		&models.Category{},
		&models.CategoryPenyakit{},
		&models.HasilDiagnosis{},
		&models.Pakar{},
		&models.Pertanyaan{},
		&models.Role{},
		&models.Sekolah{},
		&models.Students{},
		&models.Teachers{},
		&models.TestAnswer{},
		&models.TestSession{},
		&models.StudentFeedback{},
		&models.Faqs{},
		&models.BackupJob{},
		&models.StudentConsent{},      // Fitur 1A: UU PDP consent tracking
		&models.KnowledgeBaseAuditLog{}, // Fitur 3: Audit trail
	)
	//if _ != nil {
	//	log.Fatalf("Gagal migrasi database: %v", _)
	//}

	//Seeder
	seeder.SeederRole()
	seeder.SeederSekolah()
	seeder.SeederUsersAdministrator()
	seeder.SeederCategories()

	router := gin.Default()

	//router.Use(cors.New(cors.Config{
	//	AllowOrigins:     []string{"https://www.mentalhealth.web.id", "https://mentalhealth.web.id"},
	//	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	//	AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token"},
	//	ExposeHeaders:    []string{"Content-Length"},
	//	AllowCredentials: true,
	//	MaxAge:           12 * time.Hour,
	//}))

	router.GET("/api/photo/getPhotoPakar/:uid", controllers.GetPakarPhoto)
	//Index Web
	HomeRoutes := router.Group("/api/home")
	HomeRoutes.GET("/articles", controllers.GetAllArticlesHome)
	HomeRoutes.GET("/articles/:uid", controllers.GetHomeArticleByUID)
	HomeRoutes.GET("/allArticles", controllers.GetAllAriclesHome)
	HomeRoutes.GET("/thumbnail", controllers.GetThumbnailArticle)
	HomeRoutes.GET("/articles/:uid/thumbnail", controllers.GetArticleThumbnail)
	//Login Routes
	AuthRoutes := router.Group("/auth")
	AuthRoutes.POST("/registerStudents", controllers.RegisterStudents)
	AuthRoutes.POST("/loginStudents", controllers.LoginStudents)
	AuthRoutes.POST("/loginPakar", controllers.LoginPakar)
	AuthRoutes.POST("/loginAdmin", controllers.LoginAdministrator)
	AuthRoutes.POST("/loginTeacher", controllers.LoginTeachers)
	AuthRoutes.POST("/logoutAdmin", controllers.Logout)

	//Profile
	ProfileRoutes := router.Group("/api")
	ProfileRoutes.Use(middleware.RequireAuth())
	ProfileRoutes.GET("/profileStudents", controllers.GetProfileStudents)
	ProfileRoutes.GET("/profileTeachers", controllers.GetProfileTeachers)
	ProfileRoutes.GET("/profilePakars", controllers.GetProfilePakar)
	ProfileRoutes.GET("/profileAdministrator", controllers.GetProfileAdministrator)

	//Article Pakar Routes
	ArticleRoutes := router.Group("/api/article/pakar")
	ArticleRoutes.Use(middleware.RequireAuth())
	ArticleRoutes.GET("/getArticles", controllers.GetAllArticles)
	ArticleRoutes.POST("/createArticles", controllers.CreateArticle)
	ArticleRoutes.GET("/getArticleUID/:uid", controllers.GetArticleByUID)
	ArticleRoutes.PUT("/updateArticle", controllers.UpdateArticle)
	ArticleRoutes.DELETE("/deleteArticle/:uid", controllers.DeleteArticle)

	// Administrator Fitur
	// Manajemen Pengguna
	AdminRoutes := router.Group("/api/admin")
	AdminRoutes.GET("/healthcheck", controllers.GetFullDashboardData)
	AdminRoutes.GET("/getAdmin", controllers.GetAllAdministrator)
	AdminRoutes.POST("/createAdmin", controllers.CreateAdministrator)
	AdminRoutes.GET("/getAdmin/:uid", controllers.GetAdministratorByUID)
	AdminRoutes.PUT("/updateAdmin", controllers.UpdateAdministrator)
	AdminRoutes.DELETE("/deleteAdmin", controllers.DeleteAdministrator)
	// Kebijakan Retensi Data (Fitur 1B)
	AdminRoutes.GET("/retention/nearing-deletion", middleware.RequireAuth(), controllers.GetNearingDeletionStudents)
	AdminRoutes.PUT("/retention/students/:uid/status", middleware.RequireAuth(), controllers.UpdateStudentStatus)
	// Audit Trail Basis Pengetahuan (Fitur 3)
	AdminRoutes.GET("/audit-logs", middleware.RequireAuth(), controllers.GetAuditLogs)
	AdminRoutes.PUT("/sekolah/:uid/assign-pakar", middleware.RequireAuth(), controllers.AssignPakarToSchool)

	// Monitoring sekolah (drill-down: sekolah → siswa → riwayat)
	MonitoringRoutes := router.Group("/api/admin/monitoring")
	MonitoringRoutes.Use(middleware.RequireAuth())
	MonitoringRoutes.GET("/sekolah", controllers.GetSchoolMonitoring)
	MonitoringRoutes.GET("/sekolah/:npsn/siswa", controllers.GetSchoolStudents)
	MonitoringRoutes.GET("/siswa/:students_uid/riwayat", controllers.GetAdminStudentHistory)

	// Backup manual dari dashboard Admin
	BackupRoutes := router.Group("/api/admin/backup")
	BackupRoutes.Use(middleware.RequireAuth())
	BackupRoutes.POST("/trigger",              controllers.TriggerBackup)
	BackupRoutes.GET("/jobs",                  controllers.ListBackupJobs)
	BackupRoutes.GET("/jobs/:job_uid",         controllers.GetBackupJobStatus)
	BackupRoutes.GET("/download/:job_uid",     controllers.DownloadBackupFile)

	//Manajemen Role
	RoleRoutes := router.Group("/api/role")
	RoleRoutes.POST("/createRole", controllers.CreateRole)
	RoleRoutes.GET("/getRole", controllers.GetRoles)
	RoleRoutes.GET("/getRoleById/:uid", controllers.GetRoleById)
	RoleRoutes.PUT("/updateRole/:uid", controllers.UpdateRole)
	RoleRoutes.DELETE("/deleteRole/:uid", controllers.DeleteRole)
	//Manajemen Sekolah
	SchoolRoutes := router.Group("/school")
	SchoolRoutes.POST("/createSchool", controllers.CreateSchool)
	SchoolRoutes.GET("/getSchool", controllers.GetSekolah)
	SchoolRoutes.GET("/getSchoolById/:uid", controllers.GetSekolahByUID)
	SchoolRoutes.PUT("/updateSchool/:uid", controllers.UpdateSekolah)
	SchoolRoutes.DELETE("/deleteSchool/:uid", controllers.DeleteSekolah)
	SchoolRoutes.GET("/searchSchool/search", controllers.SearchSekolah)
	//Manajemen Artikel Administrator
	ArticleAdminRoutes := router.Group("/api/article/admin")
	ArticleAdminRoutes.Use(middleware.RequireAuth())
	ArticleAdminRoutes.GET("/getArticles", controllers.GetAllArticles)
	ArticleAdminRoutes.POST("/createArticles", controllers.CreateArticle)
	ArticleAdminRoutes.GET("/getArticleUID/:uid", controllers.GetArticleByUID)
	ArticleAdminRoutes.PUT("/updateArticle/:uid", controllers.UpdateArticle)
	ArticleAdminRoutes.DELETE("/deleteArticle/:uid", controllers.DeleteArticle)
	//Manajemen Users
	UsersAdminRoutes := router.Group("/api/users")
	UsersAdminRoutes.Use(middleware.RequireAuth())
	// User Students
	UsersAdminRoutes.GET("/getAllStudents", controllers.GetAllStudents)
	UsersAdminRoutes.POST("/createStudents", controllers.CreateStudents)
	UsersAdminRoutes.GET("/getStudentsById/:uid", controllers.GetStudentById)
	UsersAdminRoutes.PUT("/updateStudents/:uid", controllers.UpdateStudents)
	UsersAdminRoutes.DELETE("/deleteStudents/:uid", controllers.DeleteStudent)
	// User Pakar
	UsersAdminRoutes.GET("/getAllPakar", controllers.GetAllPakar)
	UsersAdminRoutes.POST("/createPakar", controllers.CreatePakar)
	UsersAdminRoutes.GET("/getPakarById/:uid", controllers.GetPakarByUID)
	UsersAdminRoutes.PUT("/updatePakar/:uid", controllers.UpdatePakar)
	UsersAdminRoutes.DELETE("/deletePakar/:uid", controllers.DeletePakar)
	// User Teacher
	UsersAdminRoutes.GET("/getAllTeachers", controllers.GetAllTeachers)
	UsersAdminRoutes.POST("/createTeacher", controllers.CreateTeachers)
	UsersAdminRoutes.GET("/getTeacherById/:uid", controllers.GetTeachersByUID)
	UsersAdminRoutes.PUT("/updateTeacher/:uid", controllers.UpdateTeachers)
	UsersAdminRoutes.DELETE("/deleteTeacher/:uid", controllers.DeleteTeachers)
	//Manajemen Catgeory
	CategoriesRoute := router.Group("/categories")
	CategoriesRoute.GET("/getAllCategories", controllers.GetAllCategories)

	//Pakar
	//Manajemen Artikel Pakar
	ArtikelPakarRoutes := router.Group("/api/artikelpakar")
	ArtikelPakarRoutes.Use(middleware.RequireAuth())
	ArtikelPakarRoutes.GET("/getAllArtikel", controllers.GetArticlesByAuthor)
	ArtikelPakarRoutes.GET("/getArtikelByUID/:uid", controllers.GetArticleByUID)
	ArtikelPakarRoutes.PUT("/updateArticle/:uid", controllers.UpdateArticle)
	ArtikelPakarRoutes.DELETE("/deleteArticle/:uid", controllers.DeleteArticle)

	//Manajemen Artikel Guru BK
	ArtikelGurubkRoutes := router.Group("/api/article/gurubk")
	ArtikelGurubkRoutes.Use(middleware.RequireAuth())
	ArtikelGurubkRoutes.GET("/getArticles", controllers.GetArticlesByAuthor)
	ArtikelGurubkRoutes.POST("/createArticles", controllers.CreateArticle)
	ArtikelGurubkRoutes.GET("/getArticleUID/:uid", controllers.GetArticleByUID)
	ArtikelGurubkRoutes.PUT("/updateArticle/:uid", controllers.UpdateArticle)
	ArtikelGurubkRoutes.DELETE("/deleteArticle/:uid", controllers.DeleteArticle)

	// Dashboard Pakar
	PakarDashboardRoutes := router.Group("/api/pakar/dashboard")
	PakarDashboardRoutes.Use(middleware.RequireAuth())
	PakarDashboardRoutes.GET("", controllers.GetPakarDashboardData)

	// Sekolah Binaan Pakar
	PakarBinaanRoutes := router.Group("/api/pakar")
	PakarBinaanRoutes.Use(middleware.RequireAuth())
	PakarBinaanRoutes.GET("/sekolah-binaan", controllers.GetPakarSekolahBinaan)

	//Basis Data
	//Penyakit
	PenyakitRoutes := router.Group("/api/penyakit")
	PenyakitRoutes.Use(middleware.RequireAuth())
	PenyakitRoutes.GET("/getAllPenyakits", controllers.GetAllPenyakits)
	PenyakitRoutes.GET("/getPenyakitsByUID/:uid", controllers.GetPenyakitByUID)
	PenyakitRoutes.POST("/createPenyakit", controllers.SavePenyakit)
	PenyakitRoutes.PUT("/updatePenyakit/:uid", controllers.UpdatePenyakit)
	PenyakitRoutes.DELETE("/deletePenyakit/:uid", controllers.DeletePenyakit)
	//Pertanyaan
	PertanyaanRoutes := router.Group("/api/pertanyaan")
	PertanyaanRoutes.Use(middleware.RequireAuth())
	PertanyaanRoutes.GET("/getAllPertanyaans", controllers.GetAllPertanyaans)
	PertanyaanRoutes.GET("/getPertanyaanByUID/:uid", controllers.GetPertanyaanByUID)
	PertanyaanRoutes.POST("/createPertanyaan", controllers.SavePertanyaan)
	PertanyaanRoutes.PUT("/updatePertanyaan/:uid", controllers.UpdatePertanyaan)
	PertanyaanRoutes.DELETE("/deletePertanyaan/:uid", controllers.DeletePertanyaan)
	//Aturan(Rules)
	AturanRoutes := router.Group("/api/aturan")
	AturanRoutes.Use(middleware.RequireAuth())
	AturanRoutes.GET("/getAllAturan", controllers.GetAllAturan)
	AturanRoutes.GET("/getAturan/:uid", controllers.GetAturanByUID)
	AturanRoutes.POST("/createAturan", controllers.SaveAturan)
	AturanRoutes.PUT("/updateAturan/:uid", controllers.UpdateAturan)
	AturanRoutes.DELETE("/deleteAturan/:uid", controllers.DeleteAturan)
	AturanRoutes.POST("/simulasi", controllers.SimulateRules)


	//TypeTes
	TesRoutes := router.Group("/api/tesType")
	TesRoutes.Use(middleware.RequireAuth())
	TesRoutes.GET("/getAllTypeTes", controllers.GetCategoryPenyakits)
	TesRoutes.POST("/createTypeTes", controllers.SaveCategoryPenyakit)
	TesRoutes.GET("/getTypeTesByUID/:uid", controllers.GetCategoryPenyakitsByUID)
	TesRoutes.PUT("/updateTypeTes/:uid", controllers.UpdateCategoryPenyakit)
	TesRoutes.DELETE("/deleteTypeTes/:uid", controllers.DeleteCategoryPenyakit)

	//Fitur Utama
	TesDiagnosis := router.Group("/api/diagnosis")
	TesDiagnosis.Use(middleware.RequireAuth())
	TesDiagnosis.POST("/startTes", controllers.StartTest)
	TesDiagnosis.POST("/submitTes", controllers.SubmitTest)

	// History Gurubk Routes
	GurubkHistoryRoutes := router.Group("/api/gurubk/history")
	GurubkHistoryRoutes.Use(middleware.RequireAuth())
	GurubkHistoryRoutes.GET("", controllers.GetGurubkHistory)
	GurubkHistoryRoutes.GET("/:nisn", controllers.GetGurubkHistoryDetail)
	GurubkHistoryRoutes.PATCH("/review/:id", controllers.UpdateHistoryReview)

	// Dashboard Gurubk Routes
	GurubkDashboardRoutes := router.Group("/api/gurubk/dashboard")
	GurubkDashboardRoutes.Use(middleware.RequireAuth())
	GurubkDashboardRoutes.GET("", controllers.GetGurubkDashboardData)

	// FAQ Guru BK Routes
	GurubkFaqRoutes := router.Group("/api/gurubk/faq")
	GurubkFaqRoutes.Use(middleware.RequireAuth())
	GurubkFaqRoutes.GET("/getAllFaqs", controllers.GetGuruBKFaqs)
	GurubkFaqRoutes.GET("/getFaq/:uid", controllers.GetFaqByUID)
	GurubkFaqRoutes.POST("/replyFaq/:uid", controllers.ReplyFaq)
	GurubkFaqRoutes.DELETE("/deleteFaq/:uid", controllers.DeleteFaq)

	// History Siswa Routes
	SiswaHistoryRoutes := router.Group("/api/siswa")
	SiswaHistoryRoutes.Use(middleware.RequireAuth())
	SiswaHistoryRoutes.GET("/my-history", controllers.GetStudentHistory)
	SiswaHistoryRoutes.POST("/ask-question", controllers.AskQuestion)
	SiswaHistoryRoutes.GET("/my-questions", controllers.GetStudentQuestions)
	SiswaHistoryRoutes.GET("/tren-diagnosis/:student_uid", controllers.GetStudentDiagnosisTrend)

	// Consent UU PDP Routes (Fitur 1A)
	ConsentRoutes := router.Group("/api/siswa/consent")
	ConsentRoutes.Use(middleware.RequireAuth())
	ConsentRoutes.GET("/status", controllers.GetConsentStatus)
	ConsentRoutes.POST("/submit", controllers.SubmitConsent)

	// FAQ Pakar/Admin Routes
	FaqRoutes := router.Group("/api/faq")
	FaqRoutes.Use(middleware.RequireAuth())
	FaqRoutes.GET("/getAllFaqs", controllers.GetAllFaqs)
	FaqRoutes.GET("/getFaq/:uid", controllers.GetFaqByUID)
	FaqRoutes.POST("/replyFaq/:uid", controllers.ReplyFaq)
	FaqRoutes.DELETE("/deleteFaq/:uid", controllers.DeleteFaq)

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":  "Welcome to Skripsi Backend!",
			"version":  "1.0.0",
			"Author":   "Rizky Budiarto",
			"Username": "KingRovs",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server berjalan di http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
