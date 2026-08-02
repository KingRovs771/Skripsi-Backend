package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"Skripsi-Backend/utils"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

type HomeArticleResponse struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Date      string `json:"date"`
	Author    string `json:"author"`
	Thumbnail string `json:"thumbnail"`
}

type ArticleListResponse struct {
	ArticleUID   string    `json:"article_uid"`
	JudulArticle string    `json:"judul_article"`
	Author       string    `json:"author"`
	CategoryName string    `json:"category_name"`
	Status       int64     `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type ArticleDetailResponse struct {
	ArticleUID   string          `json:"article_uid"`
	JudulArticle string          `json:"judul_article"`
	IsiArticle   string          `json:"isi_article"`
	Author       string          `json:"author"`
	Category     models.Category `json:"category"`
	Status       int             `json:"status"`
}

type ArticleInput struct {
	ArticleUID   *string `json:"article_uid"`
	JudulArticle *string `json:"judul_article"`
	IsiArticle   *string `json:"isi_article"`
	Author       *string `json:"author"`
	CategoryUID  *string `json:"category_uid"`
	Status       *int64  `json:"status"`
	Thumbnails   *string `json:"thumbnails"` // Base64 string from frontend
}

func CreateArticle(c *gin.Context) {
	judul := c.PostForm("judul_article")
	isi := c.PostForm("isi_article")
	author := c.PostForm("author")
	categoryUID := c.PostForm("category_uid")

	if judul == "" || isi == "" || categoryUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Judul, isi, dan category_uid tidak boleh kosong",
		})
		return
	}

	var categoryCount int64
	database.DB.Model(&models.Category{}).Where("category_uid = ?", categoryUID).Count(&categoryCount)
	if categoryCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Kategori tidak ditemukan!",
		})
		return
	}

	file, err := c.FormFile("thumbnails")
	var imageData []byte
	if err != nil && err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Gagal memproses file thumbnails",
		})
		return
	}

	if file != nil {
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Status":  "Error",
				"Message": "Gagal membuka file",
				"Error":   err.Error(),
			})
			return
		}
		defer src.Close()

		imageData, err = ioutil.ReadAll(src)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Status":  "Error",
				"Message": "Gagal membaca data file",
				"Error":   err.Error(),
			})
			return
		}
	}
	article := models.Article{
		JudulArticle: judul,
		IsiArticle:   isi,
		Author:       author,
		CategoryUID:  categoryUID,
		Thumbnails:   imageData,
	}

	if err := database.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal menyimpan artikel",
			"Error":   err.Error(),
		})
		return
	}

	response := gin.H{
		"Status":  "Success",
		"Message": "Artikel berhasil dibuat",
		"Data": gin.H{
			"article_uid":   article.ArticleUID,
			"judul_article": article.JudulArticle,
			"author":        article.Author,
		},
	}
	c.JSON(http.StatusCreated, response)
}

func GetHomeArticleByUID(c *gin.Context) {
	uid := c.Param("uid")

	var article models.Article
	if err := database.DB.Where("article_uid = ?", uid).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Artikel tidak ditemukan",
			"Error":   err.Error(),
		})
		return
	}

	var category models.Category
	database.DB.Where("category_uid = ?", article.CategoryUID).First(&category)

	response := ArticleDetailResponse{
		ArticleUID:   article.ArticleUID,
		JudulArticle: article.JudulArticle,
		IsiArticle:   article.IsiArticle,
		Author:       article.Author,
		Category:     category,
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Artikel",
		"Data":    response,
	})
}

func GetAllArticles(c *gin.Context) {
	var articles []models.Article
	if err := database.DB.Order("article_uid desc").Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil data artikel",
			"Error":   err.Error(),
		})
		return
	}

	if len(articles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"Status":  "Not Found",
			"Message": "Artikel Tidak Ditemukan",
			"Data":    []models.Article{},
		})
		return
	}

	categoryUIDs := make([]string, len(articles))
	for i, article := range articles {
		categoryUIDs[i] = article.CategoryUID
	}

	var categories []models.Category
	database.DB.Where("category_uid IN ?", categoryUIDs).Find(&categories)

	categoryMap := make(map[string]string)
	for _, category := range categories {
		categoryMap[category.CategoryUID] = category.NameCategory
	}

	var response []ArticleListResponse
	for _, article := range articles {
		response = append(response, ArticleListResponse{
			ArticleUID:   article.ArticleUID,
			JudulArticle: article.JudulArticle,
			Author:       article.Author,
			CategoryName: categoryMap[article.CategoryUID],
			Status:       article.Status,
			CreatedAt:    article.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Artikel",
		"data":    response,
	})
}

func GetArticleByUID(c *gin.Context) {
	uid := c.Param("uid")

	var article models.Article
	if err := database.DB.Where("article_uid = ?", uid).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Artikel tidak ditemukan",
			"Error":   err.Error(),
		})
		return
	}

	var category models.Category
	database.DB.Where("category_uid = ?", article.CategoryUID).First(&category)

	response := ArticleDetailResponse{
		ArticleUID:   article.ArticleUID,
		JudulArticle: article.JudulArticle,
		IsiArticle:   article.IsiArticle,
		Author:       article.Author,
		Category:     category,
		Status:       int(article.Status),
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Artikel",
		"Data":    response,
	})
}

func GetArticleThumbnail(c *gin.Context) {
	uid := c.Param("uid")
	var article models.Article

	if err := database.DB.Where("article_uid = ?", uid).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Artikel tidak ditemukan",
			"Error":   err.Error(),
		})
		return
	}

	if article.Thumbnails == nil || len(article.Thumbnails) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Artikel ini tidak memiliki thumbnail",
		})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", article.Thumbnails)
}
func GetAllArticlesHome(c *gin.Context) {
	var articles []models.Article

	if err := database.DB.Where("status = ?", 1).Order("created_at desc").Limit(15).Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil artikel",
			"Error":   err.Error(),
		})
		return
	}

	if len(articles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"Status":  "Success",
			"Message": "Data artikel tidak ada",
			"Data":    []HomeArticleResponse{},
		})
		return
	}

	var response []HomeArticleResponse
	for _, article := range articles {
		summary := stripHtmlTags(article.IsiArticle)
		if len(summary) > 150 {
			summary = summary[:150] + "..."
		}

		thumbnailURL := ""
		if len(article.Thumbnails) > 0 {
			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}

			baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
			thumbnailURL = fmt.Sprintf("%s/api/home/articles/%s/thumbnail", baseURL, article.ArticleUID)
		}

		response = append(response, HomeArticleResponse{
			Slug:      article.ArticleUID,
			Title:     article.JudulArticle,
			Summary:   summary,
			Date:      article.CreatedAt.Format("02 January 2006"), // Gunakan "02" agar tanggal 1-9 ada nol di depan
			Author:    article.Author,
			Thumbnail: thumbnailURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Artikel",
		"Data":    response,
	})
}

func stripHtmlTags(content string) string {
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(content, "")
}

func GetThumbnailArticle(c *gin.Context) {
	uid := c.Param("uid")
	var article models.Article

	if err := database.DB.Select("thumbnails").Where("article_uid = ?", uid).First(&article).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Deteksi tipe konten (apakah png, jpg, dsb)
	contentType := http.DetectContentType(article.Thumbnails)

	c.Data(http.StatusOK, contentType, article.Thumbnails)
}

func GetAllAriclesHome(c *gin.Context) {
	var articles []models.Article
	if err := database.DB.Where("status = ?", 1).Order("created_at desc").Limit(15).Find(&articles).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil artikel",
			"Error":   err.Error(),
		})
		return
	}

	if len(articles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"Status":  "Success",
			"Message": "Data artikel tidak ada",
			"Data":    []HomeArticleResponse{},
		})
		return
	}

	var response []HomeArticleResponse
	for _, article := range articles {
		summary := article.IsiArticle
		if len(summary) > 150 {
			summary = summary[:150] + "..."
		}

		thumbnailURL := ""
		if len(article.Thumbnails) > 0 {
			baseURL := "https://" + c.Request.Host
			thumbnailURL = fmt.Sprintf("%s/api/home/articles/%s/thumbnail", baseURL, article.ArticleUID)
		}

		response = append(response, HomeArticleResponse{
			Slug:      article.ArticleUID,
			Title:     article.JudulArticle,
			Summary:   summary,
			Date:      article.CreatedAt.Format("2 January 2006"),
			Author:    article.Author,
			Thumbnail: thumbnailURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Artikel",
		"Data":    response,
	})
}

func UpdateArticle(c *gin.Context) {
	uid := c.Param("uid")

	var artikel models.Article
	if err := database.DB.Where("article_uid = ?", uid).First(&artikel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Artikel tidak ditemukan",
			"Error":   err.Error(),
		})
		return
	}

	var inputArticle ArticleInput
	if err := c.ShouldBindJSON(&inputArticle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Data",
			"Error":   err.Error(),
		})
		return
	}

	// Update fields manually to handle pointers and special cases
	if inputArticle.JudulArticle != nil {
		artikel.JudulArticle = *inputArticle.JudulArticle
	}
	if inputArticle.IsiArticle != nil {
		artikel.IsiArticle = *inputArticle.IsiArticle
	}
	if inputArticle.Author != nil {
		artikel.Author = *inputArticle.Author
	}
	if inputArticle.CategoryUID != nil {
		artikel.CategoryUID = *inputArticle.CategoryUID
	}
	if inputArticle.Status != nil {
		artikel.Status = *inputArticle.Status
	}

	// Handle Thumbnail update if provided as base64
	if inputArticle.Thumbnails != nil && *inputArticle.Thumbnails != "" {
		// Decode base64 string
		decoded, err := base64.StdEncoding.DecodeString(*inputArticle.Thumbnails)
		if err == nil {
			artikel.Thumbnails = decoded
		}
	}

	// Save the updated article
	if err := database.DB.Save(&artikel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal memperbarui artikel",
			"Error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Update Data Artikel",
		"Data":    artikel,
	})
}

func DeleteArticle(c *gin.Context) {
	uid := c.Param("uid")

	resultArtikel := database.DB.Where("article_uid = ?", uid).Delete(&models.Article{})

	if resultArtikel.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Failed Delete Data Artikel",
		})
		return
	}

	if resultArtikel.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Error",
			"Message": "Artikel Tidak Ada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Menghapus Data Artikel",
		"Data":    resultArtikel,
	})
}
func GetArticlesByAuthor(c *gin.Context) {
	claims, err := utils.ValidateJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": "Sesi login tidak valid atau sudah kedaluwarsa",
		})
		return
	}

	var authorName string
	// Cari nama berdasarkan Tipe User
	if claims.UserType == "Pakar" {
		var pakar models.Pakar
		database.DB.Where("pakar_uid = ?", claims.ID).First(&pakar)
		authorName = pakar.NamaLengkap
	} else if claims.UserType == "Administrator" || claims.UserType == "admin" || claims.UserType == "administrator" {
		var admin models.Administrator
		database.DB.Where("admin_uid = ?", claims.ID).First(&admin)
		authorName = admin.NamaLengkap
	} else if claims.UserType == "Teachers" || claims.UserType == "teacher" || claims.UserType == "gurubk" {
		var teacher models.Teachers
		dbErr := database.DB.Where("n_ip = ?", claims.ID).First(&teacher).Error
		if dbErr != nil {
			dbErr = database.DB.Where("nip = ?", claims.ID).First(&teacher).Error
		}
		if dbErr != nil {
			dbErr = database.DB.Where("email = ?", claims.Email).First(&teacher).Error
		}
		authorName = teacher.NamaLengkap
	}

	if authorName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "Error",
			"Message": fmt.Sprintf("Sesi login tidak valid atau author tidak ditemukan (UserType: %s, ID: %s)", claims.UserType, claims.ID),
		})
		return
	}

	var articles []models.Article

	// 2. Query filter berdasarkan author (Gunakan Nama Lengkap dari JWT)
	if err := database.DB.Where("author = ?", authorName).Order("created_at desc").Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Gagal mengambil data artikel",
			"Error":   err.Error(),
		})
		return
	}

	// Jika data kosong
	if len(articles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"Status":  "Success",
			"Message": "Anda belum menulis artikel apapun",
			"data":    []interface{}{},
		})
		return
	}

	// 3. Ambil Nama Kategori (Logika Map)
	categoryUIDs := make([]string, len(articles))
	for i, article := range articles {
		categoryUIDs[i] = article.CategoryUID
	}

	var categories []models.Category
	database.DB.Where("category_uid IN ?", categoryUIDs).Find(&categories)

	categoryMap := make(map[string]string)
	for _, category := range categories {
		categoryMap[category.CategoryUID] = category.NameCategory
	}

	// 4. Mapping Data ke Map (Tanpa Merubah Model)
	var finalResponse []map[string]interface{}
	baseURL := "http://" + c.Request.Host

	for _, article := range articles {
		statusLabel := "Draft"
		if article.Status == 1 {
			statusLabel = "Publish"
		}

		thumbnailURL := ""
		if len(article.Thumbnails) > 0 {
			thumbnailURL = fmt.Sprintf("%s/api/home/articles/%s/thumbnail", baseURL, article.ArticleUID)
		}

		item := map[string]interface{}{
			"article_uid":   article.ArticleUID,
			"judul_article": article.JudulArticle,
			"category_name": categoryMap[article.CategoryUID],
			"author":        article.Author,
			"status":        article.Status,
			"status_label":  statusLabel,
			"thumbnail_url": thumbnailURL,
			"created_at":    article.CreatedAt.Format("02 January 2006"),
		}

		finalResponse = append(finalResponse, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil mendapatkan daftar artikel Anda",
		"data":    finalResponse,
	})
}
