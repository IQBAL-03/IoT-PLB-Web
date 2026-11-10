// Package model defines core domain entities and data transfer objects.
package model

// User represents an account holder in the system.
type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
}

// UserClaims represents the payload embedded into signed JWT tokens.
type UserClaims struct {
	UserID int    `json:"uid"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Exp    int64  `json:"exp"`
}

// RegisterReq defines user registration payload.
type RegisterReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyRegisterOTPReq defines payload for verifying OTP and creating the user account.
type VerifyRegisterOTPReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// LoginReq defines credentials for email/password authentication.
type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// GoogleAuthReq defines authentication data coming from Google OAuth.
type GoogleAuthReq struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// SendOTPReq defines payload for requesting an OTP code.
type SendOTPReq struct {
	Email string `json:"email"`
}



