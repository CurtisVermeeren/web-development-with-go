package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	fmt.Fprintf(w, "<h1>Welcome to my Awesome site!</h1>")

}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>To get in touch email me me@gmail.com</h1>")
}

func faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>FAQ</h1><ul><li>Is this good? Yes!</li><li>How long is this here? For now.</li><li>Is this mysterious? Perhaps.</li></ul>")
}

func renderTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	// Create templates
	t, err := template.ParseFiles("exp/hello.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		Name   string
		Date   time.Time
		Agenda map[string]string
		Money  int
	}{"Curtis", time.Now(), map[string]string{"tuesday": "Clean floors", "thursday": "Clean bathroom", "monday": "dust", "friday": "sweep"}, 5}

	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	router := mux.NewRouter()

	// Setup routes
	router.HandleFunc("/", home)
	router.HandleFunc("/contact", contact)
	router.HandleFunc("/faq", faq)
	router.HandleFunc("/template", renderTemplate)

	// Handle route not found
	router.NotFoundHandler = http.HandlerFunc(home)

	http.ListenAndServe(":3000", router)

}
