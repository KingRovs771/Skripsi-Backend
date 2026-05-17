package main

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()

	var faqs []models.Faqs
	err = database.DB.Find(&faqs).Error
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Total FAQs:", len(faqs))
	for _, f := range faqs {
		fmt.Printf("UID: %s, UserUID: %s, Question: %s\n", f.FaqsUID, f.UserUID, f.FaqPertanyaan)
	}

	// Check students table too
	var students []models.Students
	err = database.DB.Limit(5).Find(&students).Error
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nStudents (First 5):")
	for _, s := range students {
		fmt.Printf("UID: %s, Name: %s\n", s.StudentsUID, s.NamaLengkap)
	}
}
