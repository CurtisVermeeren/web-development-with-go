package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/curtisvermeeren/web-development-with-go/context"
	"github.com/curtisvermeeren/web-development-with-go/models"
	"github.com/curtisvermeeren/web-development-with-go/rand"
	"github.com/curtisvermeeren/web-development-with-go/views"
)

type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        models.UserService
}

// SignupForm represents the input fields of the sign up form page
type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// LoginForm represents the input fields of the login form page
type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// NewUsers creates and returns a Users object
func NewUsers(us models.UserService) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
	}
}

// New is used to render the form where a user can create a new account.
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	parseURLParams(r, &form)
	u.NewView.Render(w, r, form)
}

// Create is used to create a new user account.
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form SignupForm
	vd.Yield = &form

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}

	// Create the user in the db
	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	// Sign in the user
	err := u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// Login is used to process the login form when a user attempts to use an existing user
// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	// pasre login form
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	// Attempt to authenticate a user with provided credentials
	user, err := u.us.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("No user exists with that email address")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, r, vd)
		return
	}

	// Sign in the user
	err = u.signIn(w, user)
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// POST /logout
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	user := context.User(r.Context())
	token, _ := rand.RememberToken()
	user.Remember = token
	u.us.Update(user)
	http.Redirect(w, r, "/", http.StatusFound)
}

// signIn is used to sign in the given user via cookies
func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	// If no remember token exists for a user create one and update the db
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	return nil
}

// CookieTest is used to display cookies set on the current user
func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	// Get the remember token cookie
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the user from the remember token
	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, user)
}
