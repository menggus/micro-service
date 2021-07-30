package service

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// User User struct
type User struct {
	UserName     string
	HashPassword string
	Role         string
}

// NewUser return a new user
func NewUser(username, password, role string) (*User, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %s", err)
	}

	user := &User{
		UserName:     username,
		HashPassword: string(hashPassword),
		Role:         role,
	}

	return user, nil
}

// IsCorrectPassword check if the provide password is correct or not
func (user *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(password))

	return err == nil
}

// Clone create a new User
func (user *User) Clone() *User {

	return &User{
		UserName:     user.UserName,
		HashPassword: user.HashPassword,
		Role:         user.Role,
	}
}
