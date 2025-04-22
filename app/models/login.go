package models

import "github.com/golang-jwt/jwt/v4"

type CustomClaims struct {
	UserID   int    `json:"user_id"`
	UserName string `json:"user_name"`
	jwt.StandardClaims
}

type Login struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token      string `json:"token"`
	ID         int    `json:"id"`
	Username   string `json:"username"`
	RoleName   string `json:"role_name"`
	Nickname   string `json:"nickname"`
	Age        int    `json:"age"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Status     string `json:"status"`
	Sex        string `json:"sex"`
	Phone      string `json:"phone"`
	Permission string `json:"permission"`
}

type Register struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Age      string `json:"age"`
}
