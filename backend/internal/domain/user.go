package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	id           UserID
	email        string
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewUser(id UserID, email string, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		id:           id,
		email:        email,
		passwordHash: string(hash),
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func NewUserFromDB(id UserID, email string, passwordHash string, createdAt, updatedAt time.Time) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Email() string {
	return u.email
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(password))
	return err == nil
}

func GenerateUserID() UserID {
	return UserID(uuid.New())
}

func GenerateAccountID() AccountID {
	return AccountID(uuid.New())
}
