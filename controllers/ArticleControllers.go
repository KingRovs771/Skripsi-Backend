package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
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
	ArticleUID   string `json:"article_uid"`
	JudulArticle string `json:"judul_article"`
	Author       string `json:"author"`
	CategoryName string `json:"category_name"`
}

type ArticleDetailResponse struct {
	ArticleUID   string          `json:"article_uid"`
	JudulArticle string          `json:"judul_article"`
	IsiArticle   string          `json:"isi_article"`
	Author       string          `json:"author"`
	Category     models.Category `json:"category"`
}

func CreateArticle(c *gin.Context) {
	judul := c.PostForm("judul_article")
	isi := c.PostForm("isi_article")
	author := c.PostForm("author")
	categoryUID := c.PostForm("category_uid")

	if judul == "" || isi == "" || categoryUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Judul, isi, dan category_uid tidak boleh kosong"})
		return
	}

	var categoryCount int64
	database.DB.Model(&models.Category{}).Where("category_uid = ?", categoryUID).Count(&categoryCount)
	if categoryCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan!"})
		return
	}

	file, err := c.FormFile("thumbnails")
	var imageData []byte
	if err != nil && err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal memproses file thumbnails"})
		return
	}

	if file != nil {
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file"})
			return
		}
		defer src.Close()

		imageData, err = ioutil.ReadAll(src)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data file"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan artikel"})
		return
	}

	response := gin.H{
		"message": "Artikel berhasil dibuat",
		"data": gin.H{
			"article_uid":   article.ArticleUID,
			"judul_article": article.JudulArticle,
			"author":        article.Author,
		},
	}
	c.JSON(http.StatusCreated, response)
}

// GetAllArticles mengambil semua artikel dengan lookup manual.
func GetAllArticles(c *gin.Context) {
	var articles []models.Article
	if err := database.DB.Order("article_id desc").Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data artikel"})
		return
	}

	// Jika tidak ada artikel, kembalikan array kosong
	if len(articles) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []models.Article{}})
		return
	}

	// 1. Kumpulkan semua CategoryUID dari artikel
	categoryUIDs := make([]string, len(articles))
	for i, article := range articles {
		categoryUIDs[i] = article.CategoryUID
	}

	// 2. Ambil semua kategori yang relevan dalam satu query
	var categories []models.Category
	database.DB.Where("category_uid IN ?", categoryUIDs).Find(&categories)

	// 3. Buat map untuk lookup cepat: map[CategoryUID] -> NamaCategory
	categoryMap := make(map[string]string)
	for _, category := range categories {
		categoryMap[category.CategoryUID] = category.NameCategory
	}

	// 4. Buat respons dengan data yang sudah digabungkan
	var response []ArticleListResponse
	for _, article := range articles {
		response = append(response, ArticleListResponse{
			ArticleUID:   article.ArticleUID,
			JudulArticle: article.JudulArticle,
			Author:       article.Author,
			CategoryName: categoryMap[article.CategoryUID], // Ambil dari map
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetArticleByUID mengambil satu artikel dengan lookup manual.
func GetArticleByUID(c *gin.Context) {
	uid := c.Param("uid")
	var article models.Article
	if err := database.DB.Where("article_uid = ?", uid).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artikel tidak ditemukan"})
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

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetArticleThumbnail menyajikan gambar dari database.
func GetArticleThumbnail(c *gin.Context) {
	uid := c.Param("uid")
	var article models.Article

	if err := database.DB.Where("article_uid = ?", uid).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artikel tidak ditemukan"})
		return
	}

	if article.Thumbnails == nil || len(article.Thumbnails) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artikel ini tidak memiliki thumbnail"})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", article.Thumbnails)
}
func GetHomeArticles(c *gin.Context) {
	var articles []models.Article
	if err := database.DB.Order("created_at desc").Limit(10).Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil artikel"})
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
			baseURL := "rhttps://" + c.Request.Host
			thumbnailURL = fmt.Sprintf("%s/api/articles/%s/thumbnail", baseURL, article.ArticleUID)
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

	c.JSON(http.StatusOK, gin.H{"data": response})
}
