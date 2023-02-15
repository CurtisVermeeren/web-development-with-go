package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/curtisvermeeren/web-development-with-go/controllers"
	"github.com/curtisvermeeren/web-development-with-go/middleware"
	"github.com/curtisvermeeren/web-development-with-go/models"
	"github.com/curtisvermeeren/web-development-with-go/rand"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func NotFound(w http.ResponseWriter, r *http.Request) {}

func main() {
	config := LoadConfig()
	dbConfig := config.Database

	// User service
	services, err := models.NewServices(
		models.WithGorm(dbConfig.Dialect(), dbConfig.ConnectionInfo()),
		models.WithLogMode(!config.IsProd()),
		models.WithUser(config.Pepper, config.HMACKey),
		models.WithGallery(),
		models.WithImage(),
	)
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

	// Setup CSRF middleware supplied by gorilla/csrf package
	b, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}
	csrfMw := csrf.Protect(b, csrf.Secure(config.IsProd()))

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

	// Assets
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	router.PathPrefix("/assets/").Handler(assetHandler)

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

	fmt.Printf("Listening on port :%d\n", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), csrfMw(userMw.Apply(router)))

}
