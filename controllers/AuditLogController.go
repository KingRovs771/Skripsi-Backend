package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAuditLogs mengembalikan daftar log audit perubahan basis pengetahuan (paginated).
// GET /api/admin/audit-logs
func GetAuditLogs(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 15
	}
	offset := (page - 1) * limit

	query := database.DB.Model(&models.KnowledgeBaseAuditLog{})

	// Filter
	if role := c.Query("operator_role"); role != "" {
		query = query.Where("operator_role = ?", role)
	}
	if tbl := c.Query("tabel_terdampak"); tbl != "" {
		query = query.Where("tabel_terdampak = ?", tbl)
	}
	if aks := c.Query("aksi"); aks != "" {
		query = query.Where("aksi = ?", aks)
	}
	if name := c.Query("operator_name"); name != "" {
		query = query.Where("operator_name ILIKE ?", "%"+name+"%")
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal menghitung total log audit",
		})
		return
	}

	// Fetch logs
	var logs []models.KnowledgeBaseAuditLog
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil log audit",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Success",
		"Data": gin.H{
			"logs":  logs,
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
