package persistance

import (
	"authservice/src/internal/core/organizer"
	"authservice/src/pkg/utilities"
	"errors"
	"fmt"
	"log"
)

type OrganizerRepo struct {
	db *Database
}

func NewOrganizerRepo(d *Database) OrganizerRepo {
	return OrganizerRepo{
		db: d,
	}
}

func (u *OrganizerRepo) CreateUser(newUser organizer.OrgRegister) (organizer.OrgResponse, error) {
	var cid int
	var createdUser organizer.OrgResponse
	var dataCount int
	query := `select count(*) as count from users where username=$1 or email=$2`
	err := u.db.db.QueryRow(query, newUser.UserName, newUser.Email).Scan(&dataCount)
	if dataCount != 0 {
		err1 := errors.New("User or Email already exist")
		log.Println("Failed to create TempUser :", err1.Error())
		return createdUser, err1
	}
	hashPass, err := utilities.HashPassword(newUser.Password)
	if err != nil {
		fmt.Println(err, "unable to hash password")
	}

	query = "insert into users(username, email, password, profile) values($1, $2, $3, 'organizer') returning cid, created_at"
	err = u.db.db.QueryRow(query, newUser.UserName, newUser.Email, hashPass).Scan(&cid,
		&createdUser.CreatedAt)

	if err != nil {
		return organizer.OrgResponse{}, err
	}
	createdUser.Uid = cid
	return createdUser, nil
}

func (u *OrganizerRepo) GetUser(username string) (organizer.GetOrgProfile, error) {
	var newUser organizer.GetOrgProfile
	query := "select cid, username, email, created_at, password from users where username = $1 AND profile = 'organizer'"
	err := u.db.db.QueryRow(query, username).Scan(&newUser.Uid, &newUser.Username, &newUser.Email, &newUser.CreatedAt, &newUser.Password)
	if err != nil {
		return organizer.GetOrgProfile{}, err
	}
	return newUser, nil
}

func (u *OrganizerRepo) GetUserByID(id int) (organizer.OrgProfile, error) {
	var newUser organizer.OrgProfile
	query := "select cid, username, email, created_at from users where cid = $1 AND profile = 'organizer'"
	err := u.db.db.QueryRow(query, id).Scan(&newUser.Uid, &newUser.Username, &newUser.Email, &newUser.CreatedAt)
	if err != nil {
		return organizer.OrgProfile{}, err
	}
	return newUser, nil
}

func (u *OrganizerRepo) GetUsers() ([]organizer.GetOrgResponse, error) {
	var allUsers []organizer.GetOrgResponse
	query := `select cid, username from users where profile = 'organizer'`
	rows, err := u.db.db.Query(query)
	if err != nil {
		return allUsers, err
	}
	defer rows.Close()
	for rows.Next() {
		var currentUser organizer.GetOrgResponse
		err = rows.Scan(&currentUser.Uid, &currentUser.Username)
		if err != nil {
			return allUsers, err
		}
		allUsers = append(allUsers, currentUser)
	}
	return allUsers, nil
}

// Temp users methods for OTP verification flow
func (o *OrganizerRepo) CreateTempUser(newUser organizer.OrgRegister) error {
	var dataCount int
	query := `select count(*) as count from users where username=$1 or email=$2`
	err := o.db.db.QueryRow(query, newUser.UserName, newUser.Email).Scan(&dataCount)
	if dataCount != 0 {
		err1 := errors.New("User or Email already exist")
		log.Println("Failed to create TempUser :", err1.Error())
		return err1
	}

	hashPass, err := utilities.HashPassword(newUser.Password)
	if err != nil {
		log.Println("unable to hash password", err)
		return err
	}

	query = "insert into temp_users(username, email, password, profile) values($1, $2, $3, 'organizer')"
	_, err = o.db.db.Exec(query, newUser.UserName, newUser.Email, hashPass)
	if err != nil {
		return err
	}
	return nil
}

func (o *OrganizerRepo) MoveTempUserToUsers(username string) error {
	// Start transaction
	tx, err := o.db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get temp user data
	var tempUser organizer.GetOrgProfile
	query := "select username, email, password, created_at from temp_users where username = $1 AND profile = 'organizer'"
	err = tx.QueryRow(query, username).Scan(&tempUser.Username, &tempUser.Email, &tempUser.Password, &tempUser.CreatedAt)
	if err != nil {
		return err
	}

	// Insert into users table
	var cid int
	insertQuery := "insert into users(username, email, password, profile, created_at) values($1, $2, $3, 'organizer', $4) returning cid"
	err = tx.QueryRow(insertQuery, tempUser.Username, tempUser.Email, tempUser.Password, tempUser.CreatedAt).Scan(&cid)
	if err != nil {
		return err
	}

	// Delete from temp_users
	deleteQuery := "delete from temp_users where username = $1 AND profile = 'organizer'"
	_, err = tx.Exec(deleteQuery, username)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}
