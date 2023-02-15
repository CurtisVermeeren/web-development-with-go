package models

/*
UserService is an interface to manipulate the user model. It is the entrypoint to the user model.
UserService contains UserDB layers.
The UserService layers are nested so that: userService contains userValidator -> userValidator contains userGorm -> userGorm contains a hash instance and db connection.

UserDB is an interface describing methods needed to interact with the user model.

userValidator is a struct that is used for validating and normalizing data before database entry.

userGorm is a struct that is used for interacting directly with the database.
*/

import (
	"regexp"
	"strings"

	"github.com/curtisvermeeren/web-development-with-go/hash"
	"github.com/curtisvermeeren/web-development-with-go/rand"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is returned when a resource cannot be found in the database
	ErrNotFound modelError = "models: resource not found"
	// ErrIDInvalid is returned whan an invalid ID is passed to a method
	ErrIDInvalid modelError = "models: ID provided was invalid"
	// ErrPasswordIncorrect is returned when an invalid password is used when attempting to authenticate a user
	ErrPasswordIncorrect modelError = "models: incorrect password was provided"
	// ErrEmailRequired is returned when an email address is not provided when creating a user
	ErrEmailRequired modelError = "models: email address is required"
	// ErrEmailInvalid is returned when an email address does not match requirements
	ErrEmailInvalid modelError = "models: email address provided is invalid"
	// ErrEmailTaken is returned when an update or create is attempted on an email that already exists
	ErrEmailTaken modelError = "models: email address is already taken"
	// ErrPasswordTooShort is returned when a password doesn't meet a minimum length
	ErrPasswordTooShort modelError = "models: password must be at least 8 characters long"
	// ErrPasswordRequired is returned whan a create is attempted without a user password provided
	ErrPasswordRequired modelError = "models: password is required"

	// ErrRememberRequired is returned when a create or update is attemepted without a user remeber token hash
	ErrRememberRequired modelError = "models: remember token is required"
	// ErrRememberTooShort is returned when a remember token is not at least 32 bytes
	ErrRememberTooShort modelError = "models: remember token must be at least 32 bytes"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	// format model error strings for the public
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")

}

// UserService is a set of methods used to manipulate and work with the user model
type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
}

// UserService is used as an abstraction layer to the database
type userService struct {
	UserDB
	pepper string
}

// NewUserService creates a UserService object from a gorm.Db db connection
func NewUserService(db *gorm.DB, pepper, hmacKey string) UserService {
	ug := &userGorm{db}
	hmac := hash.NewHMAC(hmacKey)
	uv := newUserValidator(ug, hmac, pepper)
	return &userService{
		UserDB: uv,
	}
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

// UserDB defines methods used to interact with the users database
type UserDB interface {
	// Methods for querying a single user
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
}

// userGorm represents the database interaction layer
type userGorm struct {
	db *gorm.DB
}

// userValidator represents the data validation and normalization layer
type userValidator struct {
	UserDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
	pepper     string
}

// newUserValidator returns a userValidator object
func newUserValidator(udb UserDB, hmac hash.HMAC, pepper string) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
		pepper:     pepper,
	}
}

// userValFn type defines what is expected of a user validation function
// having all validation functions of this type allow them to be ren sequentially.
type userValFn func(*User) error

// ByID is used to find a user with matching ID
// will return ErrNotFound if no matching user is found
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ByEmail will normalize an email adreess before passing it to the next layer for querying
func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}

	err := runUserValFns(&user, uv.normalizeEmail)
	if err != nil {
		return nil, err
	}

	return uv.UserDB.ByEmail(user.Email)
}

// ByEmail is used to find a user with matching Email
// will return ErrNotFound if no matching user is found
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ByRemember will hash the remember token then call ByRemember on the subsuquent UserDB layer
func (uv *userValidator) ByRemember(token string) (*User, error) {
	// Set the remember token on a user then hash it
	user := User{
		Remember: token,
	}
	if err := runUserValFns(&user, uv.hmacRemember); err != nil {
		return nil, err
	}

	// Set the hash and pass that to the next layer
	return uv.UserDB.ByRemember(user.RememberHash)
}

// ByRemember is used to find a user with a matching remember token
// will return ErrNotFound if no matching user is found
// the method expects the token to already be hashed
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	// Search for the first user where remember_hash matches
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create will create the provided user
func (uv *userValidator) Create(user *User) error {

	// Run validation on the new user. bcrypt hash the password, set the remember token, hash the remember token
	err := runUserValFns(user,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequried,
		uv.setRememberIfUnset,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return nil
	}

	// Pass the user to the next layer
	return uv.UserDB.Create(user)
}

// Create is used to add a user to the database
func (ug *userGorm) Create(user *User) error {
	// Add user to the database
	return ug.db.Create(user).Error
}

// Update will hash a remember token is provided
func (uv *userValidator) Update(user *User) error {
	// Run validation on the new user. bcrypt hash the password, hash the remember token
	err := runUserValFns(user,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequried,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return err
	}
	// Pass the updated user to the next layer
	return uv.UserDB.Update(user)
}

// Update is used to update the provided user with all data in the provided user
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// Delete is used to delete the user with the provided ID
func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	// Check that the id of the delete user is greater than 0
	err := runUserValFns(&user, uv.idGreaterThan(0))
	if err != nil {
		return err
	}
	// Pass the id of the user to be deleted to the next layer
	return uv.UserDB.Delete(id)
}

// Delete is used to remove a user matching the provided ID
func (ug *userGorm) Delete(id uint) error {
	// Create a user with the id matching the one to be deleted
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
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
Returns ErrPasswordIncorrect if the password provided is invalid.
Returns the user if both password and email are valid.
*/
func (us *userService) Authenticate(email, password string) (*User, error) {
	// Check for user with matching email
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	// Compare the users hashed password to the provided password
	err = bcrypt.CompareHashAndPassword(
		[]byte(foundUser.PasswordHash),
		[]byte(password+us.pepper),
	)

	// Check for errors when comparing passwords
	switch err {
	case nil:
		return foundUser, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrPasswordIncorrect
	default:
		return nil, err
	}
}

// runUserValsFns runs a number of validation functions on user
// returns an error if any of the validations fail
func runUserValFns(user *User, fns ...userValFn) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

// bcryptPassword is used to hash a user's password with an application wide pepper and bycrypt, which adds salt automatically
func (uv *userValidator) bcryptPassword(user *User) error {

	// If the password hasn't changed return without hashing
	if user.Password == "" {
		return nil
	}

	// Hash the users password with bcrypt
	pwBytes := []byte(user.Password + uv.pepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

// hmacRemember is used to check if a remember token was set then hash it, otherwise return nil
func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

// setRememberIfUnset is used to check that a remeber token is set on a new user
func (uv *userValidator) setRememberIfUnset(user *User) error {
	// If remember is set then return
	if user.Remember != "" {
		return nil
	}

	// Set a remember token
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

// idGreaterThan is a closure used to check that a users id is greater than the specified value
func (uv *userValidator) idGreaterThan(n uint) userValFn {
	return userValFn(func(user *User) error {
		if user.ID <= n {
			return ErrIDInvalid
		}
		return nil
	})
}

// normalizeEmail is used to validate an email submitted
// converts the email to all lowercase
// trims whitespace in the email
func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

// requireEmail is used to validate that an email is set
func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

// emailFormat is used to validate that an email is of the required format
func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

// emailIsAvail is used to validate if an email adress is available
func (uv *userValidator) emailIsAvail(user *User) error {
	// Check if the email exists in the db
	existing, err := uv.ByEmail(user.Email)
	// Email was not found then it is available
	if err == ErrNotFound {
		return nil
	}
	// Check for other errors with the db query
	if err != nil {
		return err
	}
	// If email was found compare id of users to check if they're the same
	if user.ID != existing.ID {
		return ErrEmailTaken
	}
	return nil
}

// passwordMinLength is used to validate if a password meets the minimum length
func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// passwordRequired is used to a validate that a password is provided
func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

// passwordHashRequried is used to validate that a password is hashed
func (uv *userValidator) passwordHashRequried(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

// rememberMinBytes is used to validate that a remember token is at least 32 bytes long
func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return ErrRememberTooShort
	}
	return nil
}

// rememberHashRequired is used to validate that a remember token hash is always set
func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberRequired
	}
	return nil
}
