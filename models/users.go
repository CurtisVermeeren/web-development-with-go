package models

import (
	"errors"
	"fmt"
	"os"

	"github.com/curtisvermeeren/web-development-with-go/hash"
	"github.com/curtisvermeeren/web-development-with-go/rand"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is returned when a resource cannot be found in the database
	ErrNotFound = errors.New("models: resource not found")
	// ErrInvalidID is returned whan an invalid ID is passed to a method
	ErrInvalidID = errors.New("models: ID provided was invalid")
	// ErrInvalidPassword is returned when an invalid password is used when attempting to authenticate a user
	ErrInvalidPassword = errors.New("models: incorrect password provided")
)

// Password pepper for the application
var userPasswordPepper = os.Getenv("PASSWORDPEPPER")

// UserService is used as an abstraction layer to the database
type UserService struct {
	db   *gorm.DB
	hmac hash.HMAC
}

var hmacSecretKey = os.Getenv("SECRETHMACKEY")

// NewUserService creates a UserServiceObject using a database connectionInfo string
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, fmt.Errorf("cannot create UserService: %w", err)
	}

	// Enable Gorm logging to show statments used
	db.LogMode(true)

	// Set HMAC for hashing
	hmac := hash.NewHMAC(hmacSecretKey)

	// Return new UserService
	return &UserService{
		db:   db,
		hmac: hmac,
	}, nil
}

// Close is used to close the database connection of a UserService
func (us *UserService) Close() error {
	return us.db.Close()
}

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// ByID is used to find a user with matching ID
// will return ErrNotFound if no matching user is found
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ByEmail is used to find a user with matching Email
// will return ErrNotFound if no matching user is found
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ByRemember is used to find a user with a matching remember token
// will return ErrNotFound if no matching user is found
func (us *UserService) ByRemember(token string) (*User, error) {
	var user User
	rememberHash := us.hmac.Hash(token)
	// Search for the first user where remember_hash matches
	err := first(us.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create is used to add a user to the database
func (us *UserService) Create(user *User) error {
	// Add pepper to users pasword
	pwBytes := []byte(user.Password + userPasswordPepper)
	// GenerateFromPassword returns a salted hash
	hashedBytes, err := bcrypt.GenerateFromPassword(
		pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	// Set a remember token for the new user
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}

	// Hash the remember token and store it on the user
	user.RememberHash = us.hmac.Hash(user.Remember)

	return us.db.Create(user).Error
}

// Update is used to update the provided user with all data in the provided user
func (us *UserService) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = us.hmac.Hash(user.Remember)
	}
	return us.db.Save(user).Error
}

// Delete is used to remove a user matching the provided id
func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	// Create a user with the id matching the one to be deleted
	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

// DestructiveReset drops the user table and rebuilds it
func (us *UserService) DestructiveReset() error {
	err := us.db.DropTableIfExists(&User{}).Error
	if err != nil {
		return err
	}
	return us.AutoMigrate()
}

// AutoMigrate is used to attempt to automatically migrate the user table
func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

// first is used to query the provided gorm.DB and return the first item in dst
// will return ErrNotFound if no value is retrieved
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

/*
Authenticate is used to authenticate a user with the provided emial address and password.
Returns ErrNoFound if the email provided is invalid.
Returns ErrInvalidPassword if the password provided is invalid.
Returns the user if both password and email are valid.
*/
func (us *UserService) Authenticate(email, password string) (*User, error) {
	// Check for user with matching email
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	// Compare the users hashed password to the provided password
	err = bcrypt.CompareHashAndPassword(
		[]byte(foundUser.PasswordHash),
		[]byte(password+userPasswordPepper),
	)

	// Check for errors when comparing passwords
	switch err {
	case nil:
		return foundUser, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrInvalidPassword
	default:
		return nil, err
	}
}
