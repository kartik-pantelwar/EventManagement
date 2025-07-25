package customer

import (
	"authservice/src/internal/core/customer"
	customerservice "authservice/src/internal/usecase/customer"
	errorhandling "authservice/src/pkg/error_handling"
	pkgresponse "authservice/src/pkg/response"
	"authservice/src/pkg/utilities"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type CustomerHandler struct {
	customerService customerservice.UserService
	redisClient     *redis.Client
}

func NewCustomerHandler(usecase customerservice.UserService, redisClient *redis.Client) CustomerHandler {
	return CustomerHandler{
		customerService: usecase,
		redisClient:     redisClient,
	}
}

func (c *CustomerHandler) Register(w http.ResponseWriter, r *http.Request) {
	var newUser customer.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		errorhandling.HandleError(w, "Wrong Format Data", http.StatusBadRequest)
		return
	}

	// Store user in temp_users table
	_, err := c.customerService.RegisterUser(newUser)
	if err != nil {
		errorhandling.HandleError(w, "Unable to Register Customer", http.StatusInternalServerError)
		return
	}

	// Generate OTP
	otpStr := utilities.GenerateOtp()

	// Store OTP in Redis if available
	if c.redisClient != nil {
		err = utilities.StoreOTPInRedis(c.redisClient, newUser.UserName, otpStr)
		if err != nil {
			log.Printf("Failed to store OTP in Redis: %v", err)
		}
	}

	// For testing, let's just log the OTP (remove this in production)
	log.Printf("OTP for user %s: %s", newUser.UserName, otpStr)

	// Send OTP via email
	err = utilities.SendOTP(newUser.Email, otpStr)
	if err != nil {
		log.Printf("Failed to send OTP email: %v", err)
		// Don't fail the registration if email sending fails
		// errorhandling.HandleError(w, "Failed to send OTP email", http.StatusInternalServerError)
		// return
	} else {
		log.Printf("OTP email sent successfully to %s", newUser.Email)
	}

	// Generate a temporary registration session ID to identify user during OTP verification
	regSessionID := utilities.GenerateRegistrationToken()

	// Store the username mapping with the registration session in Redis if available
	if c.redisClient != nil {
		err = utilities.StoreRegistrationToken(c.redisClient, regSessionID, newUser.UserName)
		if err != nil {
			log.Printf("Failed to store registration session in Redis: %v", err)
		}
	}

	// Set a temporary registration session cookie (expires in 30 minutes)
	regCookie := http.Cookie{
		Name:     "reg_session",
		Value:    regSessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   1800, // 30 minutes
	}
	http.SetCookie(w, &regCookie)

	response := pkgresponse.StandardResponse{
		Status:  "SUCCESS",
		Message: "OTP sent successfully to your email. Please verify your OTP to complete registration.",
		Data: map[string]interface{}{
			"email": newUser.Email,
		},
	}
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

func (c *CustomerHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginUser customer.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&loginUser); err != nil {
		errorhandling.HandleError(w, "Wrong Format Data", http.StatusBadRequest)
		return
	}

	loginResponse, err := c.customerService.LoginUser(loginUser)
	if err != nil {
		errorhandling.HandleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	atCookie := http.Cookie{
		Name:     "at",
		Value:    loginResponse.TokenString,
		Expires:  loginResponse.TokenExpire,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}

	sessCookie := http.Cookie{
		Name:     "sess",
		Value:    loginResponse.Session.Id.String(),
		Expires:  loginResponse.Session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}
	http.SetCookie(w, &atCookie)
	http.SetCookie(w, &sessCookie)

	response := pkgresponse.StandardResponse{
		Status:  "SUCCESS",
		Message: "Customer Login Successful",
		Data: map[string]interface{}{
			"username": loginResponse.FounUser.Username,
			"user_id":  loginResponse.FounUser.Uid,
			"role":     "customer",
		},
	}
	w.Header().Set("x-user", loginResponse.FounUser.Username)
	w.Header().Set("x-userId", strconv.Itoa(loginResponse.FounUser.Uid))
	w.Header().Set("x-role", "customer")
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

func (c *CustomerHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("user").(int)
	if !ok {
		errorhandling.HandleError(w, "User Not Found in Context", http.StatusUnauthorized)
		return
	}

	registeredUser, err := c.customerService.GetUserByID(userId)
	if err != nil {
		errorhandling.HandleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := pkgresponse.StandardResponse{
		Status:  "SUCCESS",
		Message: "Customer Profile Retrieved Successfully",
		Data:    registeredUser,
	}
	w.Header().Set("x-user", registeredUser.Username)
	w.Header().Set("x-userId", strconv.Itoa(registeredUser.Uid))
	w.Header().Set("x-role", "customer")
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

func (c *CustomerHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sess")
	if err != nil {
		errorhandling.HandleError(w, "Session Cookie Not Found", http.StatusUnauthorized)
		return
	}

	tokenString, expireTime, err := c.customerService.GetJwtFromSession(cookie.Value)
	if err != nil {
		errorhandling.HandleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	atCookie := http.Cookie{
		Name:     "at",
		Value:    tokenString,
		Expires:  expireTime,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}
	http.SetCookie(w, &atCookie)

	response := pkgresponse.StandardResponse{
		Status:  "SUCCESS",
		Message: "Customer Token Refreshed Successfully",
	}
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

func (c *CustomerHandler) LogOut(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("user").(int)
	if !ok {
		errorhandling.HandleError(w, "User Not Found in Context", http.StatusUnauthorized)
		return
	}

	err := c.customerService.LogoutUser(userId)
	if err != nil {
		errorhandling.HandleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	atCookie := http.Cookie{
		Name:     "at",
		Value:    "",
		Expires:  time.Now(),
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}
	http.SetCookie(w, &atCookie)

	sessCookie := http.Cookie{
		Name:     "sess",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}
	http.SetCookie(w, &sessCookie)

	response := pkgresponse.StandardResponse{
		Status:  "SUCCESS",
		Message: "Customer Logout Successful",
	}
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

// OTP Verification - automatically gets username using registration session cookie
func (c *CustomerHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	// Parse request body - only requires OTP
	var otpRequest struct {
		OTP string `json:"otp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&otpRequest); err != nil {
		errorhandling.HandleError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate input
	if otpRequest.OTP == "" {
		errorhandling.HandleError(w, "OTP is required", http.StatusBadRequest)
		return
	}

	// Get registration session from cookie
	regCookie, err := r.Cookie("reg_session")
	if err != nil {
		log.Printf("Registration session cookie not found: %v", err)
		errorhandling.HandleError(w, "Registration session not found. Please register again.", http.StatusBadRequest)
		return
	}

	regSessionID := regCookie.Value
	if regSessionID == "" {
		errorhandling.HandleError(w, "Invalid registration session. Please register again.", http.StatusBadRequest)
		return
	}

	// Get username using registration session from Redis
	var username string

	if c.redisClient != nil {
		username, err = utilities.GetUsernameFromRegistrationToken(c.redisClient, regSessionID)
		if err != nil {
			log.Printf("Failed to get username from registration session %s: %v", regSessionID, err)
			errorhandling.HandleError(w, "Invalid or expired registration session. Please register again.", http.StatusBadRequest)
			return
		}
	} else {
		errorhandling.HandleError(w, "Registration session verification not available (Redis required)", http.StatusInternalServerError)
		return
	}

	log.Printf("Verifying OTP for username: %s, OTP: %s", username, otpRequest.OTP)

	// Verify OTP from Redis if available
	var isValid bool

	if c.redisClient != nil {
		log.Printf("Verifying OTP from Redis for user: %s", username)
		isValid, err = utilities.VerifyOTPFromRedis(c.redisClient, username, otpRequest.OTP)
		if err != nil {
			log.Printf("Redis OTP verification error for user %s: %v", username, err)
			errorhandling.HandleError(w, fmt.Sprintf("OTP verification failed: %v", err), http.StatusBadRequest)
			return
		}
		log.Printf("Redis OTP verification result for user %s: %t", username, isValid)
	} else {
		log.Printf("Redis not available, using fallback verification")
		// Fallback: For testing without Redis, accept any 6-digit OTP
		if len(otpRequest.OTP) == 6 {
			isValid = true
			log.Printf("Fallback verification: OTP length is 6, marking as valid")
		} else {
			log.Printf("Fallback verification: OTP length is not 6, marking as invalid")
		}
	}

	if !isValid {
		log.Printf("OTP verification failed for user %s with OTP %s", username, otpRequest.OTP)
		errorhandling.HandleError(w, "Invalid or expired OTP", http.StatusBadRequest)
		return
	}

	log.Printf("OTP verification successful for user: %s", username)

	// Move user from temp_users to users table and create session
	loginResponse, err := c.customerService.VerifyOTPAndLogin(username)
	if err != nil {
		log.Printf("Failed to create user after OTP verification: %v", err)
		errorhandling.HandleError(w, "Failed to complete registration", http.StatusInternalServerError)
		return
	}

	// Clean up registration session after successful verification
	if c.redisClient != nil {
		err = utilities.DeleteRegistrationToken(c.redisClient, regSessionID)
		if err != nil {
			log.Printf("Failed to delete registration session %s: %v", regSessionID, err)
			// Don't fail the request if session cleanup fails
		}
	}

	// Clear the registration session cookie
	clearCookie := http.Cookie{
		Name:     "reg_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete the cookie
	}
	http.SetCookie(w, &clearCookie)

	// Set authentication cookies
	atCookie := http.Cookie{
		Name:     "at",
		Value:    loginResponse.TokenString,
		Expires:  loginResponse.TokenExpire,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}

	sessCookie := http.Cookie{
		Name:     "sess",
		Value:    loginResponse.Session.Id.String(),
		Expires:  loginResponse.Session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}
	http.SetCookie(w, &atCookie)
	http.SetCookie(w, &sessCookie)

	response := pkgresponse.StandardResponse{
		Status:  "SUCCESS",
		Message: "OTP verified successfully. Registration completed and logged in automatically.",
		Data: map[string]interface{}{
			"username":   loginResponse.FounUser.Username,
			"user_id":    loginResponse.FounUser.Uid,
			"email":      loginResponse.FounUser.Email,
			"created_at": loginResponse.FounUser.CreatedAt,
			"role":       "customer",
		},
	}
	w.Header().Set("x-user", loginResponse.FounUser.Username)
	w.Header().Set("x-userId", strconv.Itoa(loginResponse.FounUser.Uid))
	w.Header().Set("x-role", "customer")
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}
