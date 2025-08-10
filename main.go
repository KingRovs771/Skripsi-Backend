package main

import (
	"Skripsi-Backend/controllers"
	"Skripsi-Backend/database"
	"Skripsi-Backend/middleware"
	"Skripsi-Backend/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
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
		&models.Category{},
		&models.HasilDiagnosis{},
		&models.JawabanStudents{},
		&models.JenisKuisioner{},
		&models.Kuisioner{},
		&models.Role{},
		&models.Students{},
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

	publicRoutes := router.Group("/auth")
	publicRoutes.POST("/register", controllers.RegisterStudents)
	publicRoutes.POST("/login", controllers.LoginStudents)

	protectedRoutes := router.Group("/api")
	protectedRoutes.Use(middleware.RequireAuth)
	protectedRoutes.GET("/profile", controllers.GetProfileStudents)

	ArticleRoutes := router.Group("/api/article")
	ArticleRoutes.Use(middleware.RequireAuth)
	ArticleRoutes.GET("/", controllers.GetAllArticles)
	ArticleRoutes.POST("/create", controllers.CreateArticle)

	DashboardRoutes := router.Group("/api/home")
	DashboardRoutes.GET("/articles", controllers.GetHomeArticles)

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
