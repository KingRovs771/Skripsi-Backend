package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

func CreateRole(c *gin.Context) {
	var inputRole struct {
		RoleName    string `json:"role_name" binding:"required, max=40"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&inputRole); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid input",
			"Error":   err.Error(),
		})
		return
	}

	role := models.Role{
		RoleUID:     uuid.New().String(),
		RoleName:    inputRole.RoleName,
		Description: inputRole.Description,
	}
	if err := database.DB.Create(&role).Error; err != nil {
		if database.DB.Error != nil && database.DB.Error == gorm.ErrDuplicatedKey {
			c.JSON(http.StatusConflict, gin.H{
				"Status":  "Error",
				"Message": "Role already exists",
				"Error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal server error",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  http.StatusOK,
		"Message": "Created Role Successfully",
		"Data":    role,
	})
}
func GetRoles(c *gin.Context) {
	var roles []models.Role
	if err := database.DB.Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal server error",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  http.StatusOK,
		"Message": "Get Roles Successfully",
		"Data":    roles,
	})
}

func GetRoleById(c *gin.Context) {
	uidRole := c.Param("uid")
	var role models.Role
	if err := database.DB.Where("id = ?", uidRole).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"Status":  "Error",
				"Message": "Role not found",
				"Error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal server error",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  http.StatusOK,
		"Message": "Get Role Successfully",
		"Data":    role,
	})
}

func UpdateRole(c *gin.Context) {
	uid := c.Param("uid")

	var inputRole struct {
		RoleName    *string `json:"role_name" binding:"required, max=40"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&inputRole); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid input",
			"Error":   err.Error(),
		})
		return
	}
	var role models.Role
	if err := database.DB.Where("id = ?", uid).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"Status":  "Error",
				"Message": "Role not found",
				"Error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal server error",
			"Error":   err.Error(),
		})
		return
	}
	updates := make(map[string]interface{})
	if inputRole.RoleName != nil {
		updates["role_name"] = *inputRole.RoleName
	}
	if inputRole.Description != nil {
		updates["description"] = *inputRole.Description
	}
	updates["updated_at"] = database.DB.NowFunc()

	if err := database.DB.Model(&role).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal server error",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  http.StatusOK,
		"Message": "Updated Role Successfully",
		"Data":    &role,
	})
}

func DeleteRole(c *gin.Context) {
	uid := c.Param("uid")
	var role models.Role
	if err := database.DB.Where("id = ?", uid).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"Status":  "Error",
				"Message": "Role not found",
				"Error":   err.Error(),
			})
			return
		}
	}
	if err := database.DB.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal server error",
			"Error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  http.StatusOK,
		"Message": "Deleted Role Successfully",
		"Data":    role,
	})
}
