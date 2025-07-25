package customerservice

import (
	"authservice/src/internal/adaptors/persistance"
	"authservice/src/internal/core/customer"
	"authservice/src/internal/core/session"
	"authservice/src/pkg/utilities"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	customerRepo persistance.CustomerRepo
	sessionRepo  persistance.SessionRepo
}

func NewUserService(customerRepo persistance.CustomerRepo, sessionRepo persistance.SessionRepo) UserService {
	return UserService{customerRepo: customerRepo, sessionRepo: sessionRepo}
}

// registration function definition
func (u *UserService) RegisterUser(user customer.UserRegister) (customer.UserResponse, error) {
	// Store in temp_users table first (before OTP verification)
	err := u.customerRepo.CreateTempUser(user)
	if err != nil {
		log.Printf("Error: %v", err)
		return customer.UserResponse{}, errors.New("Something Went Wrong!")
	}

	// Return a placeholder response - actual user creation happens after OTP verification
	return customer.UserResponse{
		Uid:       0, // Will be set after OTP verification
		CreatedAt: time.Now(),
	}, nil
}

// Move user from temp_users to users after OTP verification
func (u *UserService) VerifyAndCreateUser(username string) error {
	return u.customerRepo.MoveTempUserToUsers(username)
}

// Verify OTP and create user session (combines verification + login)
func (u *UserService) VerifyOTPAndLogin(username string) (LoginResponse, error) {
	// First, move user from temp_users to users
	err := u.customerRepo.MoveTempUserToUsers(username)
	if err != nil {
		return LoginResponse{}, err
	}

	// Get the newly created user
	foundUser, err := u.customerRepo.GetUser(username)
	if err != nil {
		return LoginResponse{}, err
	}

	// Generate JWT token
	tokenString, tokenExpire, err := utilities.GenerateJWT(foundUser.Uid, "customer")
	if err != nil {
		return LoginResponse{}, err
	}

	// Create session
	sessionData := session.Session{
		Id:        uuid.New(),
		Uid:       foundUser.Uid,
		TokenHash: tokenString,
		ExpiresAt: tokenExpire,
		IssuedAt:  time.Now(),
	}

	err = u.sessionRepo.CreateSession(sessionData)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		FounUser:    foundUser,
		TokenString: tokenString,
		TokenExpire: tokenExpire,
		Session:     sessionData,
	}, nil
}

type LoginResponse struct {
	FounUser    customer.GetUserProfile
	TokenString string
	TokenExpire time.Time
	Session     session.Session
}

func (u *UserService) LoginUser(requestUser customer.UserLogin) (LoginResponse, error) {
	loginResponse := LoginResponse{}

	foundUser, err := u.customerRepo.GetUser(requestUser.Username)
	if err != nil {
		log.Printf("Error: %v", err)
		return loginResponse, errors.New("Invalid Credentials")
	}

	loginResponse.FounUser = foundUser
	if err := matchPassword(requestUser, foundUser.Password); err != nil {
		log.Printf("Error: %v", err)
		return loginResponse, errors.New("Invalid Credentials")
	}
	tokenString, tokenExpire, err := utilities.GenerateJWT(foundUser.Uid, "customer")
	loginResponse.TokenString = tokenString
	loginResponse.TokenExpire = tokenExpire

	if err != nil {
		log.Printf("Error: %v", err)
		return loginResponse, errors.New("Failed to Generate Token")
	}

	session, err := utilities.GenerateSession(foundUser.Uid)
	loginResponse.Session = session
	if err != nil {
		log.Printf("Error: %v", err)
		return loginResponse, errors.New("Failed to Generate Session")
	}

	err = u.sessionRepo.CreateSession(session)
	if err != nil {
		log.Printf("Error: %v", err)
		return loginResponse, errors.New("Failed to Create Session")
	}

	return loginResponse, nil
}

func (u *UserService) GetJwtFromSession(sess string) (string, time.Time, error) {
	var tokenString string
	var tokenExpire time.Time
	session, err := u.sessionRepo.GetSession(sess)
	if err != nil {
		log.Printf("Error: %v", err)
		return tokenString, tokenExpire, errors.New("Invalid Session")
	}

	err = matchSessionToken(sess, session.TokenHash)
	if err != nil {
		log.Printf("Error: %v", err)
		return tokenString, tokenExpire, errors.New("Session Token Mismatch")
	}

	tokenString, tokenExpire, err = utilities.GenerateJWT(session.Uid, "customer")
	if err != nil {
		log.Printf("Error: %v", err)
		return tokenString, tokenExpire, errors.New("Failed to Generate Token")
	}

	return tokenString, tokenExpire, nil
}

func (u *UserService) GetUserByID(id int) (customer.UserProfile, error) {
	newUser, err := u.customerRepo.GetUserByID(id)
	if err != nil {
		log.Println("error", err)
		return customer.UserProfile{}, errors.New("User Not Found")
	}
	return newUser, nil
}

func (u *UserService) LogoutUser(id int) error {
	err := u.sessionRepo.DeleteSession(id)
	if err != nil {
		log.Printf("Error: %v", err)
		return errors.New("Failed to Logout User")
	}
	return nil
}

func matchPassword(user customer.UserLogin, password string) error {
	// !error here
	err := utilities.CheckPassword(password, user.Password)
	if err != nil {
		log.Printf("Error: %v", err)
		return fmt.Errorf("unable to match password: %v", err)
	}

	return nil
}

func matchSessionToken(id string, tokenHash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(tokenHash), []byte(id))
	if err != nil {
		log.Printf("Error: %v", err)
		fmt.Println(err, "Unable to Match Password")
	}
	return nil
}

func (u *UserService) GetAllUsers() ([]customer.GetUserProfile, error) {
	allUsers, err := u.customerRepo.GetUsers()
	if err != nil {
		log.Printf("Error: %v", err)
		return []customer.GetUserProfile{}, errors.New("Unable to Fetch Users")
	}
	return allUsers, nil
}
