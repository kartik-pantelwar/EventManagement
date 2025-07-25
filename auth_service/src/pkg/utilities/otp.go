package utilities

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	// "math/big"
	"net/smtp"
	"time"

	"github.com/go-redis/redis/v8"
	"gopkg.in/gomail.v2"
)

// GenerateOtp generates a random 6-digit OTP
func GenerateOtp() string {
	// Use crypto/rand for better security
	// max := big.NewInt(900000) // 999999 - 100000 = 900000
	// n, err := rand.Int(rand.Reader, max)
	// if err != nil {
	// 	log.Printf("Error generating OTP: %v", err)
	// 	// Fallback to less secure method
	// 	return 100000 + int(time.Now().UnixNano()%900000)
	// }
	// return int(n.Int64()) + 100000 // Ensure 6 digits (100000-999999)
	otpInt := rand.Intn(900000) + 100000
	otp := strconv.Itoa(otpInt)
	return otp

}

// StoreOTPInRedis stores OTP in Redis with expiration
func StoreOTPInRedis(client *redis.Client, username string, otp string) error {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", username)

	// Store OTP with 10 minutes expiration
	err := client.Set(ctx, key, otp, 10*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to store OTP in Redis: %w", err)
	}

	log.Printf("OTP stored for user: %s", username)
	return nil
}

// VerifyOTPFromRedis verifies OTP from Redis
func VerifyOTPFromRedis(client *redis.Client, username string, otp string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", username)

	log.Printf("Verifying OTP for key: %s, provided OTP: %s", key, otp)

	storedOTP, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("OTP not found or expired for key: %s", key)
		return false, fmt.Errorf("OTP expired or not found")
	} else if err != nil {
		log.Printf("Redis error for key %s: %v", key, err)
		return false, fmt.Errorf("failed to get OTP from Redis: %w", err)
	}

	log.Printf("Stored OTP for key %s: %s", key, storedOTP)

	// Compare OTPs
	isValid := storedOTP == otp
	log.Printf("OTP comparison result for key %s: %t", key, isValid)

	// Delete OTP after verification attempt (one-time use) - only if verification was successful
	if isValid {
		client.Del(ctx, key)
		log.Printf("OTP deleted for key: %s (successful verification)", key)
	} else {
		log.Printf("OTP not deleted for key: %s (failed verification)", key)
	}

	return isValid, nil
}

// SendOTP sends OTP via email
func SsendOTP(email, otp string) error {
	// SMTP configuration - these should come from config
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	smtpUsername := "your-email@gmail.com"
	smtpPassword := "your-app-password"

	// Email content
	from := smtpUsername
	to := []string{email}
	subject := "Your OTP for Registration"
	body := fmt.Sprintf(`
Dear User,

Your OTP for registration is: %s

This OTP will expire in 10 minutes.

Best regards,
Auth Service Team
`, otp)

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, email, subject, body)

	// SMTP authentication
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	// Send email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	log.Printf("OTP sent to email: %s", email)
	return nil
}

func SendOTP(toEmail string, otp string) error {

	m := gomail.NewMessage()
	m.SetHeader("From", "kartik.pantelwar@gmail.com")
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Verification Mail")
	// newOtp := fmt.Sprintf("Your OTP is - %s", otp)
	// m.SetBody("text/html", newOtp)
	m.SetBody("text/html", "<h3>Your OTP is:</h3><p><b>"+otp+"</b></p>")

	d := gomail.NewDialer("smtp.gmail.com", 587, "kartik.pantelwar@gmail.com", "hlzrebsdwbfdunho")

	return d.DialAndSend(m)
}

// GenerateRegistrationToken generates a unique registration token for temporary user identification
func GenerateRegistrationToken() string {
	// Generate a random token using timestamp and random number
	timestamp := time.Now().UnixNano()
	randomNum := rand.Intn(10000)
	token := fmt.Sprintf("reg_%d_%d", timestamp, randomNum)
	return token
}

// StoreRegistrationToken stores the mapping between registration token and username in Redis
func StoreRegistrationToken(client *redis.Client, token string, username string) error {
	ctx := context.Background()
	key := fmt.Sprintf("reg_token:%s", token)

	// Store token with 30 minutes expiration (longer than OTP for multiple attempts)
	err := client.Set(ctx, key, username, 30*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to store registration token in Redis: %w", err)
	}

	log.Printf("Registration token stored for user: %s", username)
	return nil
}

// GetUsernameFromRegistrationToken retrieves username using registration token from Redis
func GetUsernameFromRegistrationToken(client *redis.Client, token string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("reg_token:%s", token)

	log.Printf("Getting username for token: %s", key)

	username, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("Registration token not found or expired for key: %s", key)
		return "", fmt.Errorf("registration token expired or not found")
	} else if err != nil {
		log.Printf("Redis error for key %s: %v", key, err)
		return "", fmt.Errorf("failed to get username from Redis: %w", err)
	}

	log.Printf("Username found for token %s: %s", key, username)
	return username, nil
}

// DeleteRegistrationToken removes the registration token after successful verification
func DeleteRegistrationToken(client *redis.Client, token string) error {
	ctx := context.Background()
	key := fmt.Sprintf("reg_token:%s", token)

	err := client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Failed to delete registration token %s: %v", key, err)
		return err
	}

	log.Printf("Registration token deleted: %s", key)
	return nil
}
