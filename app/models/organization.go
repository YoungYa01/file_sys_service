package models

type Organization struct {
	ID          int             `json:"id"`
	OrgName     string          `json:"org_name"`
	ParentId    int             `json:"parent_id"`
	OrgLogo     string          `json:"org_logo"`
	Leader      string          `json:"leader"`
	Status      int             `json:"status"`
	Sort        int             `json:"sort"`
	Description string          `json:"description"`
	Children    []*Organization `json:"children,omitempty" gorm:"-"` // 忽略数据库映射
	Users       []*User         `json:"users" gorm:"-"`
}
