package persistance

import (
	"authservice/src/internal/core/customer"
	"authservice/src/pkg/utilities"
	"errors"
	"fmt"
	"log"
)

type CustomerRepo struct {
	db *Database
}

func NewCustomerRepo(d *Database) CustomerRepo {
	return CustomerRepo{
		db: d,
	}
}

func (c *CustomerRepo) CreateUser(newUser customer.UserRegister) (customer.UserResponse, error) {
	var cid int
	var createdUser customer.UserResponse
	// query := "insert into users(username, email, password, work_location, balance) values($1, $2, $3, $4, $5) returning cid"
	hashPass, err := utilities.HashPassword(newUser.Password)
	if err != nil {
		fmt.Println(err, "unable to hash password")
	}

	query := "insert into users(username, email, password, profile) values($1, $2, $3, 'customer') returning cid, created_at"
	err = c.db.db.QueryRow(query, newUser.UserName, newUser.Email, hashPass).Scan(&cid,
		&createdUser.CreatedAt)

	if err != nil {
		return customer.UserResponse{}, err
	}
	createdUser.Uid = cid
	return createdUser, nil
}

func (u *CustomerRepo) GetUser(username string) (customer.GetUserProfile, error) {
	var newUser customer.GetUserProfile
	query := "select cid, username, email, created_at, password from users where username = $1 AND profile = 'customer'"
	err := u.db.db.QueryRow(query, username).Scan(&newUser.Uid, &newUser.Username, &newUser.Email, &newUser.CreatedAt, &newUser.Password)
	if err != nil {
		return customer.GetUserProfile{}, err
	}
	return newUser, nil
}

func (u *CustomerRepo) GetUserByID(id int) (customer.UserProfile, error) {
	var newUser customer.UserProfile
	query := "select cid, username, email, created_at from users where cid = $1 AND profile = 'customer'"
	err := u.db.db.QueryRow(query, id).Scan(&newUser.Uid, &newUser.Username, &newUser.Email, &newUser.CreatedAt)
	if err != nil {
		return customer.UserProfile{}, err
	}
	return newUser, nil
}

func (u *CustomerRepo) GetUsers() ([]customer.GetUserProfile, error) {
	var allUsers []customer.GetUserProfile
	query := `select cid, username from users where profile = 'customer'`
	rows, err := u.db.db.Query(query)
	if err != nil {
		return allUsers, err
	}
	defer rows.Close()
	for rows.Next() {
		var currentUser customer.GetUserProfile
		err = rows.Scan(&currentUser.Uid, &currentUser.Username)
		if err != nil {
			return allUsers, err
		}
		allUsers = append(allUsers, currentUser)
	}
	return allUsers, nil
}

// Temp users methods for OTP verification flow
func (c *CustomerRepo) CreateTempUser(newUser customer.UserRegister) error {
	var dataCount int
	query := `select count(*) as count from users where username=$1 or email=$2`
	err := c.db.db.QueryRow(query, newUser.UserName, newUser.Email).Scan(&dataCount)
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

	query = "insert into temp_users(username, email, password, profile) values($1, $2, $3, 'customer')"
	_, err = c.db.db.Exec(query, newUser.UserName, newUser.Email, hashPass)
	if err != nil {
		return err
	}
	return nil
}

func (c *CustomerRepo) MoveTempUserToUsers(username string) error {
	// Start transaction
	tx, err := c.db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get temp user data
	var tempUser customer.GetUserProfile
	query := "select username, email, password, created_at from temp_users where username = $1 AND profile = 'customer'"
	err = tx.QueryRow(query, username).Scan(&tempUser.Username, &tempUser.Email, &tempUser.Password, &tempUser.CreatedAt)
	if err != nil {
		return err
	}

	// Insert into users table
	var cid int
	insertQuery := "insert into users(username, email, password, profile, created_at) values($1, $2, $3, 'customer', $4) returning cid"
	err = tx.QueryRow(insertQuery, tempUser.Username, tempUser.Email, tempUser.Password, tempUser.CreatedAt).Scan(&cid)
	if err != nil {
		return err
	}

	// Delete from temp_users
	deleteQuery := "delete from temp_users where username = $1 AND profile = 'customer'"
	_, err = tx.Exec(deleteQuery, username)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}
