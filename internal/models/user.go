package models

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const UserTableName = "users"

// User represents a registered user.
type User struct {
	Model
	Username string `json:"username"`
	Password string `json:"-"`
}

// GetUserByUsername returns the User for the provided username.
func GetUserByUsername(username string) (*User, error) {
	db := DB.Table(UserTableName)
	db.Where("username = ? ", username)

	var user User
	if err := db.Find(&user).Error; err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, ErrRecordNotFound
	}
	return &user, nil
}

// GetUserByID returns the User for the provided ID.
func GetUserByID(id uint) (*User, error) {
	var user User
	db := DB.Table(UserTableName)
	if err := db.Where("id = ?", id).Find(&user).Error; err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, ErrRecordNotFound
	}
	return &user, nil
}

// CreateUser creates a user with the provided username and password.
func CreateUser(username, password string) (*User, error) {
	hashedPassword, err := GetPasswordHash(password)
	if err != nil {
		return nil, err
	}
	user := &User{
		Username: username,
		Password: hashedPassword,
	}
	db := DB.Table(UserTableName)
	if err := db.Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, ErrUniqueConstraintViolation
		}
		return nil, err
	}
	return user, nil
}

func GetPasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
