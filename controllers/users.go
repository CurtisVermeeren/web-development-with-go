package controllers

import (
	"fmt"
	"net/http"

	"github.com/curtisvermeeren/web-development-with-go/views"
)

type Users struct {
	NewView *views.View
}

// SignupForm represents the input fields of the sign up form page
type SignupForm struct {
	Email    string `schema: "email"`
	Password string `schema: "password"`
}

// NewUsers creates and returns a Users object
func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "users/new"),
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
	err := parseForm(r, &form)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(w, "Email is", form.Email)
	fmt.Fprintln(w, "Password is", form.Password)
}
