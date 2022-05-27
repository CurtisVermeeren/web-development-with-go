package main

import (
	"net/http"

	"github.com/curtisvermeeren/web-development-with-go/controllers"
	"github.com/gorilla/mux"
)

// must is a helper function that panics when an error is reached
func must(err error) {
	if err != nil {
		panic((err))
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {}

func main() {
	router := mux.NewRouter()

	// Setup Controlelrs
	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers()

	// Setup routes
	router.Handle("/", staticController.Home).Methods("GET")
	router.Handle("/contact", staticController.Contact).Methods("GET")
	router.Handle("/faq", staticController.Faq).Methods("GET")
	router.HandleFunc("/signup", usersController.New).Methods("GET")
	router.HandleFunc("/signup", usersController.Create).Methods("POST")

	router.NotFoundHandler = staticController.Home

	http.ListenAndServe(":3000", router)

}
