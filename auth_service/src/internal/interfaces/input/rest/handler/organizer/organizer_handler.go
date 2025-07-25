package organizer

import (
	"authservice/src/internal/core/organizer"
	organizerservice "authservice/src/internal/usecase/organizer"
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

type OrganizerHandler struct {
	organizerService organizerservice.UserService
	redisClient      *redis.Client
}

func NewOrganizerHandler(usecase organizerservice.UserService, redisClient *redis.Client) OrganizerHandler {
	return OrganizerHandler{
		organizerService: usecase,
		redisClient:      redisClient,
	}
}

func (o *OrganizerHandler) Register(w http.ResponseWriter, r *http.Request) {
	var newOrg organizer.OrgRegister
	if err := json.NewDecoder(r.Body).Decode(&newOrg); err != nil {
		errorhandling.HandleError(w, "Wrong Format Data", http.StatusBadRequest)
		return
	}

	// Store organizer in temp_users table
	_, err := o.organizerService.RegisterUser(newOrg)
	if err != nil {
		errorhandling.HandleError(w, "Unable to Register Organizer", http.StatusInternalServerError)
		return
	}

	// Generate OTP
	otpStr := utilities.GenerateOtp()

	// Store OTP in Redis if available
	if o.redisClient != nil {
		err = utilities.StoreOTPInRedis(o.redisClient, newOrg.UserName, otpStr)
		if err != nil {
			log.Printf("Failed to store OTP in Redis: %v", err)
		}
	}

	// For testing, let's just log the OTP (remove this in production)
	log.Printf("OTP for organizer %s: %s", newOrg.UserName, otpStr)

	// Send OTP via email
	err = utilities.SendOTP(newOrg.Email, otpStr)
	if err != nil {
		log.Printf("Failed to send OTP email: %v", err)
		// Don't fail the registration if email sending fails
		// errorhandling.HandleError(w, "Failed to send OTP email", http.StatusInternalServerError)
		// return
	} else {
		log.Printf("OTP email sent successfully to %s", newOrg.Email)
	}

	// Generate a temporary registration session ID to identify organizer during OTP verification
	regSessionID := utilities.GenerateRegistrationToken()

	// Store the username mapping with the registration session in Redis if available
	if o.redisClient != nil {
		err = utilities.StoreRegistrationToken(o.redisClient, regSessionID, newOrg.UserName)
		if err != nil {
			log.Printf("Failed to store registration session in Redis: %v", err)
		}
	}

	// Set a temporary registration session cookie (expires in 30 minutes)
	regCookie := http.Cookie{
		Name:     "org_reg_session",
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
			"email": newOrg.Email,
		},
	}
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

func (o *OrganizerHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginOrg organizer.OrgLogin
	if err := json.NewDecoder(r.Body).Decode(&loginOrg); err != nil {
		errorhandling.HandleError(w, "Wrong Format Data", http.StatusBadRequest)
		return
	}

	loginResponse, err := o.organizerService.LoginUser(loginOrg)
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
		Message: "Organizer Login Successful",
		Data: map[string]interface{}{
			"username": loginResponse.FounUser.Username,
			"user_id":  loginResponse.FounUser.Uid,
			"role":     "organizer",
		},
	}
	w.Header().Set("x-user", loginResponse.FounUser.Username)
	w.Header().Set("x-userId", strconv.Itoa(loginResponse.FounUser.Uid))
	w.Header().Set("x-role", "organizer")
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

func (o *OrganizerHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("user").(int)
	if !ok {
		errorhandling.HandleError(w, "User Not Found in Context", http.StatusUnauthorized)
		return
	}

	registeredOrg, err := o.organizerService.GetUserByID(userId)
	if err != nil {
		errorhandling.HandleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := pkgresponse.StandardResponse{
		Status:  "SUCCESS",
		Message: "Organizer Profile Retrieved Successfully",
		Data:    registeredOrg,
	}
	w.Header().Set("x-user", registeredOrg.Username)
	w.Header().Set("x-userId", strconv.Itoa(registeredOrg.Uid))
	w.Header().Set("x-role", "organizer")
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

func (o *OrganizerHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sess")
	if err != nil {
		errorhandling.HandleError(w, "Session Cookie Not Found", http.StatusUnauthorized)
		return
	}

	tokenString, expireTime, err := o.organizerService.GetJwtFromSession(cookie.Value)
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
		Message: "Organizer Token Refreshed Successfully",
	}
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

func (o *OrganizerHandler) LogOut(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("user").(int)
	if !ok {
		errorhandling.HandleError(w, "User Not Found in Context", http.StatusUnauthorized)
		return
	}

	err := o.organizerService.LogoutUser(userId)
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
		Message: "Organizer Logout Successful",
	}
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}

// OTP Verification - automatically gets username using registration session cookie
func (o *OrganizerHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
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
	regCookie, err := r.Cookie("org_reg_session")
	if err != nil {
		log.Printf("Organizer registration session cookie not found: %v", err)
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

	if o.redisClient != nil {
		username, err = utilities.GetUsernameFromRegistrationToken(o.redisClient, regSessionID)
		if err != nil {
			log.Printf("Failed to get username from registration session %s: %v", regSessionID, err)
			errorhandling.HandleError(w, "Invalid or expired registration session. Please register again.", http.StatusBadRequest)
			return
		}
	} else {
		errorhandling.HandleError(w, "Registration session verification not available (Redis required)", http.StatusInternalServerError)
		return
	}

	log.Printf("Verifying OTP for organizer username: %s, OTP: %s", username, otpRequest.OTP)

	// Verify OTP from Redis if available
	var isValid bool

	if o.redisClient != nil {
		log.Printf("Verifying OTP from Redis for organizer: %s", username)
		isValid, err = utilities.VerifyOTPFromRedis(o.redisClient, username, otpRequest.OTP)
		if err != nil {
			log.Printf("Redis OTP verification error for organizer %s: %v", username, err)
			errorhandling.HandleError(w, fmt.Sprintf("OTP verification failed: %v", err), http.StatusBadRequest)
			return
		}
		log.Printf("Redis OTP verification result for organizer %s: %t", username, isValid)
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
		log.Printf("OTP verification failed for organizer %s with OTP %s", username, otpRequest.OTP)
		errorhandling.HandleError(w, "Invalid or expired OTP", http.StatusBadRequest)
		return
	}

	log.Printf("OTP verification successful for organizer: %s", username)

	// Move organizer from temp_users to users table and create session
	loginResponse, err := o.organizerService.VerifyOTPAndLogin(username)
	if err != nil {
		log.Printf("Failed to create organizer after OTP verification: %v", err)
		errorhandling.HandleError(w, "Failed to complete registration", http.StatusInternalServerError)
		return
	}

	// Clean up registration session after successful verification
	if o.redisClient != nil {
		err = utilities.DeleteRegistrationToken(o.redisClient, regSessionID)
		if err != nil {
			log.Printf("Failed to delete registration session %s: %v", regSessionID, err)
			// Don't fail the request if session cleanup fails
		}
	}

	// Clear the registration session cookie
	clearCookie := http.Cookie{
		Name:     "org_reg_session",
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
			"role":       "organizer",
		},
	}
	w.Header().Set("x-user", loginResponse.FounUser.Username)
	w.Header().Set("x-userId", strconv.Itoa(loginResponse.FounUser.Uid))
	w.Header().Set("x-role", "organizer")
	pkgresponse.WriteResponse(w, http.StatusOK, response)
}
