package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	ID         int            `json:"id"`
	Username   string         `json:"username"`
	Password   string         `json:"password"`
	RoleID     uint           `json:"role_id"`                       // 改为关联Role表的ID
	Role       Role           `gorm:"foreignKey:RoleID" json:"role"` // 添加关联关系
	Age        int            `json:"age"`
	Email      string         `json:"email"`
	Phone      string         `json:"phone"`
	Sex        string         `json:"sex"`
	Avatar     string         `json:"avatar"`
	Status     string         `json:"status"`
	Permission string         `json:"permission"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

func (u User) TableName() string {
	return "users"
}
