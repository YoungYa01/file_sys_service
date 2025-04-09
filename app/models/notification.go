package models

import "time"

type Notification struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Founder   int       `json:"founder"`
	Pinned    string    `json:"pinned" gorm:"default:0"`
}

func (n *Notification) TableName() string {
	return "notification"
}
