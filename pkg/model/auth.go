// Package model defines core domain entities and data transfer objects.
package model

// User represents an account holder in the system.
type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	AuthProvider string `json:"auth_provider"`
	IsVerified   bool   `json:"is_verified"`
	CreatedAt    string `json:"created_at"`
}

// TrustedDevice represents an authenticated hardware/browser fingerprint.
type TrustedDevice struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	DeviceToken string `json:"device_token"`
	DeviceName  string `json:"device_name"`
	CreatedAt   string `json:"created_at"`
	LastUsedAt  string `json:"last_used_at"`
}

// OTPRecord represents an issued verification code stored in the database.
type OTPRecord struct {
	ID        int
	Email     string
	CodeHash  string
	Purpose   string
	Metadata  string
	Attempts  int
	ExpiresAt string
	IsUsed    bool
	CreatedAt string
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
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DeviceToken string `json:"device_token"`
	DeviceName  string `json:"device_name"`
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
	Email       string `json:"email"`
	Password    string `json:"password"`
	DeviceToken string `json:"device_token"`
	DeviceName  string `json:"device_name"`
}

// GoogleAuthReq defines authentication data coming from Google OAuth.
type GoogleAuthReq struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	DeviceToken string `json:"device_token"`
	DeviceName  string `json:"device_name"`
	Action      string `json:"action,omitempty"` // "login" or "register"
}

// VerifyOTPReq defines payload for verifying an OTP code.
type VerifyOTPReq struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	Purpose     string `json:"purpose"` // "registration", "login_new_device", "google_registration"
	DeviceToken string `json:"device_token"`
	DeviceName  string `json:"device_name"`
}

// SendOTPReq defines payload for requesting an OTP code.
type SendOTPReq struct {
	Email string `json:"email"`
}

// ResendOTPReq defines payload for requesting a new OTP code.
type ResendOTPReq struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}
