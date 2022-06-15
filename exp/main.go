package main

import (
	"fmt"
	"log"
	"os"

	"github.com/curtisvermeeren/web-development-with-go/models"
	_ "github.com/lib/pq"
)

func main() {
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

	userService.DestructiveReset()

	newUser := models.User{
		Name:  "Micheal Scott",
		Email: "micheal@office.com",
	}
	if err := userService.Create(&newUser); err != nil {
		log.Fatal(err)
	}

	userFound, err := userService.ByID(1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(userFound)

	userFoundEmail, err := userService.ByEmail("micheal@office.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(userFoundEmail)

	newUser.Name = "Updated Name!!"
	if err := userService.Update(&newUser); err != nil {
		log.Fatal(err)
	}

	userFound, err = userService.ByEmail("micheal@office.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(userFound)

	if err := userService.Delete(userFound.ID); err != nil {
		log.Fatal(err)
	}

	_, err = userService.ByID(userFound.ID)
	if err != models.ErrNotFound {
		log.Fatal("user was not deleted")
	}
}
