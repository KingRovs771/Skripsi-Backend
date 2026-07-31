package utils

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WriteAuditLog merekam riwayat perubahan basis pengetahuan ke tabel knowledge_base_audit_logs.
func WriteAuditLog(c *gin.Context, tableName string, recordID string, action string, before interface{}, after interface{}) {
	operatorUID, existsUID := c.Get("user_uid")
	operatorType, existsType := c.Get("user_type")

	if !existsUID || !existsType {
		log.Println("[Audit Log] Error: User context tidak ditemukan di Context Gin")
		return
	}

	uidStr := operatorUID.(string)
	typeStr := operatorType.(string)

	var name string
	if typeStr == "admin" {
		var admin models.Administrator
		if err := database.DB.Where("admin_uid = ?", uidStr).First(&admin).Error; err == nil {
			name = admin.NamaLengkap
		} else {
			name = "Administrator"
		}
	} else if typeStr == "pakar" {
		var pakar models.Pakar
		if err := database.DB.Where("pakar_uid = ?", uidStr).First(&pakar).Error; err == nil {
			name = pakar.NamaLengkap
		} else {
			name = "Pakar"
		}
	} else {
		name = "Unknown (" + typeStr + ")"
	}

	var beforeJSON, afterJSON *string

	if before != nil {
		b, err := json.Marshal(before)
		if err == nil {
			s := string(b)
			beforeJSON = &s
		}
	}

	if after != nil {
		a, err := json.Marshal(after)
		if err == nil {
			s := string(a)
			afterJSON = &s
		}
	}

	audit := models.KnowledgeBaseAuditLog{
		LogUID:            uuid.New().String(),
		OperatorUID:       uidStr,
		OperatorName:      name,
		OperatorRole:      typeStr,
		TabelTerdampak:    tableName,
		RecordIdTerdampak: recordID,
		Aksi:              action,
		DataSebelum:       beforeJSON,
		DataSesudah:       afterJSON,
		CreatedAt:         time.Now(),
	}

	if err := database.DB.Create(&audit).Error; err != nil {
		log.Printf("[Audit Log] Gagal menyimpan audit log ke DB: %v\n", err)
	}
}
