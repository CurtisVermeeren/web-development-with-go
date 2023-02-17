package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func (c PostgresConfig) Dialect() string {
	return "postgres"
}

func (c PostgresConfig) ConnectionInfo() string {
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Name)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.Name)
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "password",
		Name:     "db",
	}
}

type Config struct {
	Port     int
	Env      string
	Pepper   string
	HMACKey  string
	Database PostgresConfig
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

func DefaultConfig() Config {
	return Config{
		Port:     8080,
		Env:      "dev",
		Pepper:   "secret-random-string",
		HMACKey:  "secret-hmac-key",
		Database: DefaultPostgresConfig(),
	}
}

func LoadConfig() Config {
	// Database setup
	host := os.Getenv("POSTGRES_HOST")
	port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		log.Fatal(err)
	}
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	dbConfig := PostgresConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     dbname,
	}

	// Get the hmacSecretKey from env
	hmacSecretKey := os.Getenv("SECRETHMACKEY")

	// Password pepper for the application
	userPasswordPepper := os.Getenv("PASSWORDPEPPER")

	config := Config{
		Port:     8080,
		Env:      "dev",
		Pepper:   userPasswordPepper,
		HMACKey:  hmacSecretKey,
		Database: dbConfig,
	}

	return config
}
