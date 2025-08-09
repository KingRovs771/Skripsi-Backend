package main

import (
	"Skripsi-Backend/controllers"
	"Skripsi-Backend/database"
	"Skripsi-Backend/middleware"
	"Skripsi-Backend/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()
	err = database.DB.AutoMigrate(&models.Students{})
	if err != nil {
		log.Fatalf("Gagal migrasi database: %v", err)
	}

	router := gin.Default()
	publicRoutes := router.Group("/auth")
	publicRoutes.POST("/register", controllers.RegisterStudents)
	publicRoutes.POST("/login", controllers.LoginStudents)

	protectedRoutes := router.Group("/api")
	protectedRoutes.Use(middleware.RequireAuth)
	protectedRoutes.GET("/profile", controllers.GetProfileStudents)

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
