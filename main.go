package main

import (
	"net/http"

	"github.com/curtisvermeeren/web-development-with-go/views"
	"github.com/gorilla/mux"
)

var homeView *views.View
var contactView *views.View
var faqView *views.View

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(homeView.Render(w, nil))
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(contactView.Render(w, nil))
}

func faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(faqView.Render(w, nil))
}

// must is a helper function that panics when an error is reached
func must(err error) {
	if err != nil {
		panic((err))
	}
}

func main() {
	router := mux.NewRouter()

	// Setup views
	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")
	faqView = views.NewView("bootstrap", "views/faq.gohtml")

	// Setup routes
	router.HandleFunc("/", home)
	router.HandleFunc("/contact", contact)
	router.HandleFunc("/faq", faq)

	// Handle route not found
	router.NotFoundHandler = http.HandlerFunc(home)

	http.ListenAndServe(":3000", router)

}
