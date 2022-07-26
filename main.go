package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/curtisvermeeren/web-development-with-go/controllers"
	"github.com/curtisvermeeren/web-development-with-go/models"
	"github.com/gorilla/mux"
)

func NotFound(w http.ResponseWriter, r *http.Request) {}

func main() {

	// Database setup
	host := "host.docker.internal"
	port := 5432
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	userService, err := models.NewUserService(psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer userService.Close()

	// Used to reset the DB in development
	// userService.DestructiveReset()

	// Create the database schema
	userService.AutoMigrate()

	// Create a new router
	router := mux.NewRouter()

	// Setup Controlelrs
	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(userService)

	// Page routes
	router.Handle("/", staticController.Home).Methods("GET")
	router.Handle("/contact", staticController.Contact).Methods("GET")
	router.Handle("/faq", staticController.Faq).Methods("GET")
	router.HandleFunc("/cookietest", usersController.CookieTest).Methods("GET")
	// Signup routes
	router.HandleFunc("/signup", usersController.New).Methods("GET")
	router.HandleFunc("/signup", usersController.Create).Methods("POST")
	// Login routes
	router.Handle("/login", usersController.LoginView).Methods("GET")
	router.HandleFunc("/login", usersController.Login).Methods("POST")

	router.NotFoundHandler = staticController.Home

	http.ListenAndServe(":3000", router)

}
