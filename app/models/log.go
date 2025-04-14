package models

import "time"

type Log struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserId    int       `json:"user_id"`
	UserName  string    `json:"user_name"`
	Ip        string    `json:"ip"`
	Os        string    `json:"os"`
	Params    string    `json:"params"` // json格式
	Method    string    `json:"method"`
	ApiUrl    string    `json:"api_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Browser   string    `json:"browser"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
}
