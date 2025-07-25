package organizer

import "time"

type Org struct {
	UserID int
}

type OrgRegister struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Otp      int    `json:"otp"`
}

type OrgProfile struct {
	Uid       int       `json:"uid"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type OrgLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
	OTP      int    `json:"otp"`
}

type OrgResponse struct {
	Uid       int       `json:"uid"`
	CreatedAt time.Time `json:"created_at"`
}

type GetOrgProfile struct {
	Uid       int       `json:"uid"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Password  string    `json:"password"`
}

type GetOrgResponse struct {
	Uid      int    `json:"uid"`
	Username string `json:"username"`
}
