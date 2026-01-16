package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Article struct {
	ArticleId    int64     `gorm:"primaryKey;uniqueIndex" json:"article_id"`
	ArticleUID   string    `gorm:"type:varchar(255)" json:"article_uid"`
	CategoryUID  string    `gorm:"type:varchar(255)" json:"category_uid"`
	JudulArticle string    `gorm:"type:varchar(255)" json:"judul_article"`
	IsiArticle   string    `gorm:"type:text" json:"isi_article"`
	Author       string    `gorm:"type:varchar(255)" json:"author"`
	Status       int64     `gorm:"type:int" json:"status"`
	Thumbnails   []byte    `gorm:"type:bytea;not null" json:"thumbnails"`
	CreatedAt    time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdateAt     time.Time `gorm:"type:timestamp" json:"update_at"`
}

func (Article *Article) BeforeCreate(tx *gorm.DB) error {

	UniqueId, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	Article.ArticleUID = UniqueId.String()
	return nil
}
