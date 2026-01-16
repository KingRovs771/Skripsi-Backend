package main

import (
	"Skripsi-Backend/controllers"
	"Skripsi-Backend/database"
	"Skripsi-Backend/middleware"
	"Skripsi-Backend/models"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()
	err = database.DB.AutoMigrate(
		&models.Administrator{},
		&models.Article{},
		&models.Aturan{},
		&models.Category{},
		&models.CategoryPenyakit{},
		&models.HasilDiagnosis{},
		&models.Pakar{},
		&models.Penyakit{},
		&models.Pertanyaan{},
		&models.Role{},
		&models.Sekolah{},
		&models.Students{},
		&models.Teachers{},
		&models.TestAnswer{},
		&models.TestSession{},
	)
	if err != nil {
		log.Fatalf("Gagal migrasi database: %v", err)
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"}, // Ganti dengan domain frontend Anda
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//Index Web
	HomeRoutes := router.Group("/api/home")
	HomeRoutes.GET("/articles", controllers.GetHomeArticles)
	HomeRoutes.GET("/articles/:uid", controllers.GetHomeArticleByUID)
	HomeRoutes.GET("/allArticles", controllers.GetAllAriclesHome)

	//Login Routes
	AuthRoutes := router.Group("/auth")
	AuthRoutes.POST("/registerStudents", controllers.RegisterStudents)
	AuthRoutes.POST("/loginStudents", controllers.LoginStudents)
	AuthRoutes.POST("/loginPakar")
	AuthRoutes.POST("/loginAdmin")
	AuthRoutes.POST("/loginTeacher")

	//Profile Students
	protectedRoutes := router.Group("/api")
	protectedRoutes.Use(middleware.RequireAuth)
	protectedRoutes.GET("/profileStudents", controllers.GetProfileStudents)

	//Article Pakar Routes
	ArticleRoutes := router.Group("/api/article/pakar")
	ArticleRoutes.Use(middleware.RequireAuth)
	ArticleRoutes.GET("/getArticles", controllers.GetAllArticles)
	ArticleRoutes.POST("/createArticles", controllers.CreateArticle)
	ArticleRoutes.GET("/getArticleUID/:uid", controllers.GetArticleByUID)
	ArticleRoutes.PUT("/updateArticle", controllers.UpdateArticle)
	ArticleRoutes.DELETE("/deleteArticle/:uid", controllers.DeleteArticle)

	// Administrator Fitur
	// Manajemen Pengguna
	AdminRoutes := router.Group("/api/admin")
	AdminRoutes.GET("/getAdmin", controllers.GetAllAdministrator)
	AdminRoutes.POST("/createAdmin", controllers.CreateAdministrator)
	AdminRoutes.GET("/getAdmin/:uid", controllers.GetAdministratorByUID)
	AdminRoutes.PUT("/updateAdmin", controllers.UpdateAdministrator)
	AdminRoutes.DELETE("/deleteAdmin", controllers.DeleteAdministrator)
	//Manajemen Role
	RoleRoutes := router.Group("/api/role")
	RoleRoutes.POST("/createRole")
	RoleRoutes.GET("/getRole")
	RoleRoutes.GET("/getRoleById/:uid")
	RoleRoutes.PUT("/updateRole")
	RoleRoutes.DELETE("/deleteRole/:uid")
	//Manajemen Sekolah
	SchoolRoutes := router.Group("/school")
	SchoolRoutes.POST("/createSchool")
	SchoolRoutes.GET("/getSchool")
	SchoolRoutes.GET("/getSchoolById/:uid")
	SchoolRoutes.PUT("/updateSchool")
	SchoolRoutes.DELETE("/deleteSchool/:uid")
	//Manajemen Artikel Administrator
	ArticleAdminRoutes := router.Group("/api/article/admin")
	ArticleAdminRoutes.Use(middleware.RequireAuth)
	ArticleAdminRoutes.GET("/getArticles", controllers.GetAllArticles)
	ArticleAdminRoutes.POST("/createArticles", controllers.CreateArticle)
	ArticleAdminRoutes.GET("/getArticleUID/:uid", controllers.GetArticleByUID)
	ArticleAdminRoutes.PUT("/updateArticle", controllers.UpdateArticle)
	ArticleAdminRoutes.DELETE("/deleteArticle/:uid", controllers.DeleteArticle)
	//Manajemen Users
	UsersAdminRoutes := router.Group("/api/users")
	UsersAdminRoutes.Use(middleware.RequireAuth)
	// User Students
	UsersAdminRoutes.GET("/getAllStudents")
	UsersAdminRoutes.POST("/createStudents")
	UsersAdminRoutes.GET("/getStudentsById/:uid")
	UsersAdminRoutes.PUT("/updateStudents")
	UsersAdminRoutes.DELETE("/deleteStudents/:uid")
	// User Pakar
	UsersAdminRoutes.GET("/getAllPakar")
	UsersAdminRoutes.POST("/createPakar")
	UsersAdminRoutes.GET("/getPakarById/:uid")
	UsersAdminRoutes.PUT("/updatePakar")
	UsersAdminRoutes.DELETE("/deletePakar/:uid")
	// User Teacher
	UsersAdminRoutes.GET("/getAllTeachers")
	UsersAdminRoutes.POST("/createTeacher")
	UsersAdminRoutes.GET("/getTeacherById/:uid")
	UsersAdminRoutes.PUT("/updateTeacher")
	UsersAdminRoutes.DELETE("/deleteTeacher/:uid")

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
