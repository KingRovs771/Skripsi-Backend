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
		&models.Gejala{},
		&models.HasilDiagnosis{},
		&models.Kuisioner{},
		&models.OpsiJawaban{},
		&models.Pakar{},
		&models.Penyakit{},
		&models.Pertanyaan{},
		&models.Role{},
		&models.Sekolah{},
		&models.Students{},
		&models.Teachers{},
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
	AuthRoutes.POST("/register", controllers.RegisterStudents)
	AuthRoutes.POST("/login", controllers.LoginStudents)
	AuthRoutes.POST("/loginPakar")
	AuthRoutes.POST("/loginAdmin")
	AuthRoutes.POST("/loginTeacher")

	//Profile Students
	protectedRoutes := router.Group("/api")
	protectedRoutes.Use(middleware.RequireAuth)
	protectedRoutes.GET("/profileStudents", controllers.GetProfileStudents)

	//Article Pakar Routes
	ArticleRoutes := router.Group("/api/article")
	ArticleRoutes.Use(middleware.RequireAuth)
	ArticleRoutes.GET("/getArticles", controllers.GetAllArticles)
	ArticleRoutes.POST("/createArticles", controllers.CreateArticle)
	ArticleRoutes.GET("/getArticleUID/:uid", controllers.GetArticleByUID)
	ArticleRoutes.PUT("/updateArticle", controllers.UpdateArticle)
	ArticleRoutes.DELETE("/deleteArticle/:uid", controllers.DeleteArticle)

	//Administrator Route
	AdminRoutes := router.Group("/api/admin")
	AdminRoutes.GET("/getAdmin", controllers.GetAllAdministrator)
	AdminRoutes.POST("/createAdmin", controllers.CreateAdministrator)
	AdminRoutes.GET("/getAdmin/:uid", controllers.GetAdministratorByUID)
	AdminRoutes.PUT("/updateAdmin", controllers.UpdateAdministrator)
	AdminRoutes.DELETE("/deleteAdmin", controllers.DeleteAdministrator)

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
