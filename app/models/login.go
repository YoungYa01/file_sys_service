package models

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token      string `json:"token"`
	ID         int    `json:"id"`
	Username   string `json:"username"`
	RoleName   string `json:"role_name"`
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
