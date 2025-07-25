package customer

import "time"

type Users struct {
	UserID int
}

type UserRegister struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	// Otp      int    `json:"otp"`
}

type UserProfile struct {
	Uid       int       `json:"uid"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	Uid       int       `json:"uid"`
	CreatedAt time.Time `json:"created_at"`
}

type GetUserProfile struct {
	Uid       int       `json:"uid"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Password  string    `json:"password"`
}

type GetUserResponse struct {
	Uid      int    `json:"uid"`
	Username string `json:"username"`
}
