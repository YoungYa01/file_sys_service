package models

import (
	"time"
)

type Carousel struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	CreatedAt   time.Time `json:"-" gorm:"autoCreateTime"` // 自动创建时间
	UpdatedAt   time.Time `json:"-" gorm:"autoUpdateTime"` // 自动更新时间
}

func (c *Carousel) TableName() string {
	return "carousels"
}
