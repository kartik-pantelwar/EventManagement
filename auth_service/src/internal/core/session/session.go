package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Id        uuid.UUID `json:"id"`
	Uid       int       `json:"uid"`
	TokenHash string    `json:"tokenthash"`
	ExpiresAt time.Time `json:"expiresat"`
	IssuedAt  time.Time `json:"issuedat"`
}
