package mailer

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"strings"

	"plantcare-iot/pkg/config"
)

// GenerateOTP generates a secure 6-digit numeric OTP string.
func GenerateOTP() string {
	maxVal := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return "123456"
	}
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// SendOTP delivers an OTP verification email to the user.
func SendOTP(cfg config.Config, toEmail, otpCode string) error {
	toEmail = strings.TrimSpace(toEmail)
	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPass == "" {
		log.Printf("[DEV MODE OTP] SMTP belum disetel. Kode OTP untuk %s adalah: %s\n", toEmail, otpCode)
		return nil
	}

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)

	subject := "Subject: Kode Verifikasi Akun PlantCare IoT\r\n"
	fromHeader := fmt.Sprintf("From: %s\r\n", cfg.SMTPFrom)
	toHeader := fmt.Sprintf("To: %s\r\n", toEmail)
	mimeHeader := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n"

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; background-color: #F0FDF4; margin: 0; padding: 20px; }
    .container { max-width: 500px; margin: 0 auto; background: #ffffff; border-radius: 12px; padding: 32px; border: 1px solid #DCFCE7; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
    .header { text-align: center; margin-bottom: 24px; }
    .brand { color: #166534; font-size: 24px; font-weight: bold; margin: 0; }
    .tagline { color: #15803D; font-size: 14px; margin-top: 4px; }
    .code-box { background: #DCFCE7; border-radius: 8px; text-align: center; padding: 20px; margin: 28px 0; border: 1px dashed #22C55E; }
    .code { font-size: 36px; font-weight: 800; letter-spacing: 8px; color: #14532D; margin: 0; }
    .notice { font-size: 13px; color: #4B5563; line-height: 1.6; }
    .footer { text-align: center; font-size: 12px; color: #9CA3AF; margin-top: 28px; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1 class="brand">PlantCare IoT</h1>
      <p class="tagline">Sistem Penyiraman Tanaman Otomatis</p>
    </div>
    <p>Halo,</p>
    <p>Gunakan kode OTP berikut untuk masuk atau memverifikasi akun PlantCare IoT Anda:</p>
    <div class="code-box">
      <div class="code">%s</div>
    </div>
    <p class="notice">
      [PENTING] Kode ini berlaku selama <strong>5 menit</strong>.<br>
      Jangan bagikan kode ini kepada siapa pun demi keamanan perangkat Anda.
    </p>
    <div class="footer">
      PlantCare IoT - Project Lintas Bidang Informatika PPLG
    </div>
  </div>
</body>
</html>`, otpCode)

	msg := []byte(subject + fromHeader + toHeader + mimeHeader + htmlBody)

	return smtp.SendMail(addr, auth, cfg.SMTPUser, []string{toEmail}, msg)
}
