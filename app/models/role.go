package models

type Role struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      string `json:"status"`
	Permission  string `json:"permission"`
}

func (r Role) TableName() string {
	return "roles"
}
