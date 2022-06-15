package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/curtisvermeeren/web-development-with-go/models"
	"github.com/curtisvermeeren/web-development-with-go/views"
)

type Users struct {
	NewView *views.View
	us      *models.UserService
}

// SignupForm represents the input fields of the sign up form page
type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// NewUsers creates and returns a Users object
func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "users/new"),
		us:      us,
	}
}

// New is used to render the form where a user can create a new account.
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	err := u.NewView.Render(w, nil)
	if err != nil {
		panic(err)
	}
}

// Create is used to create a new user account.
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {

	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		log.Fatal(err)
	}

	user := models.User{
		Name:  form.Name,
		Email: form.Email,
	}

	if err := u.us.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "User is", user)
}
