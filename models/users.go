package models

import (
	"errors"
	"fmt"
	"os"

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
	db *gorm.DB
}

// NewUserService creates a UserServiceObject using a database connectionInfo string
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, fmt.Errorf("cannot create UserService: %w", err)
	}
	// Enable Gorm logging to show statments used
	db.LogMode(true)
	return &UserService{
		db: db,
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
	return us.db.Create(user).Error
}

// Update is used to update the provided user with all data in the provided user
func (us *UserService) Update(user *User) error {
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
