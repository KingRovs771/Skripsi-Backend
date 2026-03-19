package main

import (
	"Skripsi-Backend/controllers"
	"Skripsi-Backend/database"
	"Skripsi-Backend/middleware"
	"Skripsi-Backend/models"
	"Skripsi-Backend/seeder"
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
	database.ConnectRedis()
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

	//Seeder
	seeder.SeederRole()
	seeder.SeederSekolah()
	seeder.SeederUsersAdministrator()
	seeder.SeederCategories()

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
	AdminRoutes.GET("/getAdmin", controllers.GetAllAdministrator)
	AdminRoutes.POST("/createAdmin", controllers.CreateAdministrator)
	AdminRoutes.GET("/getAdmin/:uid", controllers.GetAdministratorByUID)
	AdminRoutes.PUT("/updateAdmin", controllers.UpdateAdministrator)
	AdminRoutes.DELETE("/deleteAdmin", controllers.DeleteAdministrator)
	//Manajemen Role
	RoleRoutes := router.Group("/api/role")
	RoleRoutes.POST("/createRole", controllers.CreateRole)
	RoleRoutes.GET("/getRole", controllers.GetRoles)
	RoleRoutes.GET("/getRoleById/:uid", controllers.GetRoleById)
	RoleRoutes.PUT("/updateRole", controllers.UpdateRole)
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
