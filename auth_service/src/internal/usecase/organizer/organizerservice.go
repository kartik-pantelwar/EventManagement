package organizerservice

import (
	"authservice/src/internal/adaptors/persistance"
	"authservice/src/internal/core/organizer"
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
	organizerRepo persistance.OrganizerRepo
	sessionRepo   persistance.SessionRepo
}

func NewUserService(organizerRepo persistance.OrganizerRepo, sessionRepo persistance.SessionRepo) UserService {
	return UserService{organizerRepo: organizerRepo, sessionRepo: sessionRepo}
}

// registration function definition
func (u *UserService) RegisterUser(user organizer.OrgRegister) (organizer.OrgResponse, error) {
	// Store in temp_users table first (before OTP verification)
	err := u.organizerRepo.CreateTempUser(user)
	if err != nil {
		log.Printf("Error: %v", err)
		return organizer.OrgResponse{}, errors.New("Something Went Wrong!")
	}

	// Return a placeholder response - actual user creation happens after OTP verification
	return organizer.OrgResponse{
		Uid:       0, // Will be set after OTP verification
		CreatedAt: time.Now(),
	}, nil
}

// Move user from temp_users to users after OTP verification
func (u *UserService) VerifyAndCreateUser(username string) error {
	return u.organizerRepo.MoveTempUserToUsers(username)
}

// Verify OTP and create organizer session (combines verification + login)
func (u *UserService) VerifyOTPAndLogin(username string) (LoginResponse, error) {
	// First, move organizer from temp_users to users
	err := u.organizerRepo.MoveTempUserToUsers(username)
	if err != nil {
		return LoginResponse{}, err
	}

	// Get the newly created organizer
	foundUser, err := u.organizerRepo.GetUser(username)
	if err != nil {
		return LoginResponse{}, err
	}

	// Generate JWT token
	tokenString, tokenExpire, err := utilities.GenerateJWT(foundUser.Uid, "organizer")
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
	FounUser    organizer.GetOrgProfile
	TokenString string
	TokenExpire time.Time
	Session     session.Session
}

func (u *UserService) LoginUser(requestUser organizer.OrgLogin) (LoginResponse, error) {
	loginResponse := LoginResponse{}

	foundUser, err := u.organizerRepo.GetUser(requestUser.Username)
	if err != nil {
		log.Printf("Error: %v", err)
		return loginResponse, errors.New("Invalid Credentials")
	}

	loginResponse.FounUser = foundUser
	if err := matchPassword(requestUser, foundUser.Password); err != nil {
		log.Printf("Error: %v", err)
		return loginResponse, errors.New("Invalid Credentials")
	}
	tokenString, tokenExpire, err := utilities.GenerateJWT(foundUser.Uid, "organizer")
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

	tokenString, tokenExpire, err = utilities.GenerateJWT(session.Uid, "organizer")
	if err != nil {
		log.Printf("Error: %v", err)
		return tokenString, tokenExpire, errors.New("Failed to Generate Token")
	}

	return tokenString, tokenExpire, nil
}

func (u *UserService) GetUserByID(id int) (organizer.OrgProfile, error) {
	newUser, err := u.organizerRepo.GetUserByID(id)
	if err != nil {
		log.Println("error", err)
		return organizer.OrgProfile{}, errors.New("User Not Found")
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

func matchPassword(user organizer.OrgLogin, password string) error {
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

func (u *UserService) GetAllUsers() ([]organizer.GetOrgResponse, error) {
	allUsers, err := u.organizerRepo.GetUsers()
	if err != nil {
		log.Printf("Error: %v", err)
		return []organizer.GetOrgResponse{}, errors.New("Unable to Fetch Users")
	}
	return allUsers, nil
}
