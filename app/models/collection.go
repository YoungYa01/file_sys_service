package models

import "time"

type Collection struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	FileType        string    `json:"file_type"`
	Access          string    `json:"access"`
	AccessPwd       string    `json:"access_pwd"`
	FileNumber      int       `json:"file_number"`
	Founder         int       `json:"founder"`
	Status          int       `json:"status" gorm:"default:1"`
	Pinned          int       `json:"pinned"`
	SubmittedNumber int       `json:"submitted_number"`
	TotalNumber     int       `json:"total_number"`
	Submitters      []int     `json:"submitters"`
	Reviewers       []int     `json:"reviewers"`
	EndTime         string    `json:"end_time"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type CollectionCreator struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	FileType        string    `json:"file_type"`
	Access          string    `json:"access"`
	AccessPwd       string    `json:"access_pwd"`
	FileNumber      int       `json:"file_number"`
	Founder         int       `json:"founder"`
	Status          int       `json:"status" gorm:"default:1"`
	Pinned          int       `json:"pinned"`
	SubmittedNumber int       `json:"submitted_number"`
	TotalNumber     int       `json:"total_number"`
	EndTime         time.Time `json:"end_time"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	// 添加关联关系配置
	Submitters []CollectionSubmitter `json:"submitters" gorm:"foreignKey:CollectionID;references:ID"`
	Reviewers  []CollectionReviewer  `json:"reviewers" gorm:"foreignKey:CollectionID;references:ID"`
}

func (c *CollectionCreator) TableName() string {
	return "collections"
}

// 提交者关联表
type CollectionSubmitter struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	CollectionID uint      `json:"collection_id"`
	UserID       uint      `json:"user_id"`
	UserName     string    `json:"user_name"`
	TaskStatus   uint      `json:"task_status" gorm:"size:20;default:1"`
	ReviewStatus uint      `json:"review_status"`
	ReviewTime   time.Time `json:"review_time"`
	SubmitTime   time.Time `json:"submit_time" gorm:"default:''"`
	FilePath     string    `json:"file_path"`
	FileName     string    `json:"file_name"`
	Recommend    string    `json:"recommend"`
	Sort         int       `json:"sort" gorm:"default:1"`
}

func (cs *CollectionSubmitter) TableName() string {
	return "collection_submitters"
}

// 审核者关联表（带顺序）
type CollectionReviewer struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	CollectionID uint   `json:"collection_id"`
	UserID       uint   `json:"user_id"`
	UserName     string `json:"user_name"`
	ReviewOrder  int    `json:"review_order" gorm:"default:1"` // 审核顺序（1=第一审核人）
}

func (cr *CollectionReviewer) TableName() string {
	return "collection_reviewers"
}

type TCFile struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
}

type TCSubmitter struct {
	CollectionID uint     `json:"collection_id"`
	UserID       uint     `json:"user_id"`
	File         []TCFile `json:"file"`
}
