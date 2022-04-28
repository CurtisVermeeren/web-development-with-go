package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/curtisvermeeren/web-development-with-go/views"
	"github.com/gorilla/mux"
)

var homeView *views.View
var contactView *views.View

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	err := homeView.Template.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := contactView.Template.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>FAQ</h1><ul><li>Is this good? Yes!</li><li>How long is this here? For now.</li><li>Is this mysterious? Perhaps.</li></ul>")
}

func main() {
	router := mux.NewRouter()

	// Setup views
	homeView = views.NewView("views/home.gohtml")
	contactView = views.NewView("views/contact.gohtml")

	// Setup routes
	router.HandleFunc("/", home)
	router.HandleFunc("/contact", contact)
	router.HandleFunc("/faq", faq)

	// Handle route not found
	router.NotFoundHandler = http.HandlerFunc(home)

	http.ListenAndServe(":3000", router)

}
