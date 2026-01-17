package controllers

import (
	"Skripsi-Backend/database"
	"Skripsi-Backend/models"
	"fmt"
	"io/ioutil"
	"net/http"
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
	CreatedAt    time.Time `json:"created_at"`
}

type ArticleDetailResponse struct {
	ArticleUID   string          `json:"article_uid"`
	JudulArticle string          `json:"judul_article"`
	IsiArticle   string          `json:"isi_article"`
	Author       string          `json:"author"`
	Category     models.Category `json:"category"`
}

type ArticleInput struct {
	ArticleUID   string `json:"article_uid"`
	JudulArticle string `json:"judul_article"`
	IsiArticle   string `json:"isi_article"`
	Author       string `json:"author"`
	CategoryUID  string `json:"category_uid"`
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
func GetHomeArticles(c *gin.Context) {
	var articles []models.Article
	if err := database.DB.Order("created_at desc").Limit(5).Find(&articles).Error; err != nil {

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

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Berhasil Mendapatkan Data Artikel",
		"Data":    response,
	})
}

func GetAllAriclesHome(c *gin.Context) {
	var articles []models.Article
	if err := database.DB.Order("created_at desc").Limit(15).Find(&articles).Error; err != nil {

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "UID Tidak Ditemukan",
			"Error":   err.Error(),
		})
		return
	}

	var inputArticle ArticleInput
	if err := c.ShouldBindBodyWithJSON(&inputArticle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Error",
			"Message": "Invalid Input Data",
			"Error":   err.Error(),
		})
		return
	}

	if err := database.DB.Model(&artikel).Updates(inputArticle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":  "Error",
			"Message": "Internal Server Error",
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
