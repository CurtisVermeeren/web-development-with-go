package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/curtisvermeeren/web-development-with-go/controllers"
	"github.com/curtisvermeeren/web-development-with-go/middleware"
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

	// User service
	services, err := models.NewServices(psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer services.Close()

	// Used to reset the DB in development
	// services.DestructiveReset()

	// Create the database schema
	services.AutoMigrate()

	// Create a new router
	router := mux.NewRouter()

	// Setup Controlelrs
	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User)
	galleriesController := controllers.NewGalleries(services.Gallery, services.Image, router)

	// Setup middleware
	userMw := middleware.User{
		UserService: services.User,
	}

	requireUserMw := middleware.RequireUser{}

	// Apply middleware
	createGallery := requireUserMw.ApplyFn(galleriesController.Create)
	editGallery := requireUserMw.ApplyFn(galleriesController.Edit)
	updateGallery := requireUserMw.ApplyFn(galleriesController.Update)
	deleteGallery := requireUserMw.ApplyFn(galleriesController.Delete)
	indexGallery := requireUserMw.ApplyFn(galleriesController.Index)
	uploadGallery := requireUserMw.ApplyFn(galleriesController.ImageUpload)
	deleteImage := requireUserMw.ApplyFn(galleriesController.ImageDelete)

	// Image routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

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
	// Gallery routes
	router.Handle("/galleries/new", galleriesController.New).Methods("GET")
	router.HandleFunc("/galleries", createGallery).Methods("POST")
	router.HandleFunc("/galleries/{id:[0-9]+}", galleriesController.Show).Methods("GET").Name(controllers.ShowGallery)
	router.HandleFunc("/galleries/{id:[0-9]+}/edit", editGallery).Methods("GET").Name(controllers.EditGallery)
	router.HandleFunc("/galleries/{id:[0-9]+}/update", updateGallery).Methods("POST")
	router.HandleFunc("/galleries/{id:[0-9]+}/delete", deleteGallery).Methods("POST")
	router.Handle("/galleries", indexGallery).Methods("GET").Name(controllers.IndexGalleries)
	router.HandleFunc("/galleries/{id:[0-9]+}/images", uploadGallery).Methods("POST")
	router.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", deleteImage).Methods("POST")

	router.NotFoundHandler = staticController.Home

	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", userMw.Apply(router))

}
