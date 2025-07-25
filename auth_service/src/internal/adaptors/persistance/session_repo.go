package persistance

import (
	"authservice/src/internal/core/session"
	"fmt"
)

type SessionRepo struct {
	db *Database
}

func NewSessionRepo(d *Database) SessionRepo {
	return SessionRepo{db: d}
}

func (u *SessionRepo) CreateSession(session session.Session) error {
	_, err := u.db.db.Exec("INSERT INTO sessions (id, user_id, token_hash, expires_at, issued_at) VALUES($1, $2, $3, $4, $5) ON CONFLICT(user_id) DO UPDATE SET id= EXCLUDED.id, token_hash=EXCLUDED.token_hash, expires_at=EXCLUDED.expires_at, issued_at=EXCLUDED.issued_at", session.Id, session.Uid, session.TokenHash, session.ExpiresAt, session.IssuedAt)
	if err != nil {
		return err
	}
	fmt.Println("Session inserted into db", session.Id)
	return nil
}

func (u *SessionRepo) GetSession(id string) (session.Session, error) {
	var newSess session.Session
	query := "select id, user_id, token_hash, expires_at, issued_at from sessions where id=$1"
	err := u.db.db.QueryRow(query, id).Scan(&newSess.Id, &newSess.Uid, &newSess.TokenHash, &newSess.ExpiresAt, &newSess.IssuedAt)
	if err != nil {
		return session.Session{}, err
	}
	return newSess, nil
}

func (u *SessionRepo) GetSessionByUid(uid string) (session.Session, error) {
	var newSess session.Session
	query := "select id, user_id, token_hash, expires_at, issued_at from sessions where user_id = $1"
	err := u.db.db.QueryRow(query, uid).Scan(&newSess.Id, &newSess.Uid, &newSess.TokenHash, &newSess.ExpiresAt, &newSess.IssuedAt)
	if err != nil {
		return session.Session{}, err
	}
	return newSess, nil
}

func (u *SessionRepo) DeleteSession(uid int) error {
	query := "delete from sessions where user_id=$1"
	_, err := u.db.db.Query(query, uid)
	if err != nil {
		return err
	}
	return nil
}

// GetUserRole gets the role of a user (organizer or customer)
func (u *SessionRepo) GetUserRole(userID int) (string, error) {
	var role string

	// First check if user is an organizer
	query := "SELECT 'organizer' FROM organizers WHERE uid = $1"
	err := u.db.db.QueryRow(query, userID).Scan(&role)
	if err == nil {
		return role, nil
	}

	// Then check if user is a customer
	query = "SELECT 'customer' FROM customers WHERE uid = $1"
	err = u.db.db.QueryRow(query, userID).Scan(&role)
	if err == nil {
		return role, nil
	}

	return "", fmt.Errorf("user role not found")
}
