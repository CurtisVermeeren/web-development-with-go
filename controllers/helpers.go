package controllers

import (
	"net/http"

	"github.com/gorilla/schema"
)

// parseForm takes an http.Request and parses and data from the form
// Data parsed is stored in the destination interface
func parseForm(r *http.Request, destination interface{}) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Use the gorilla schema package for decoding into a destination interface
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(destination, r.PostForm)
	if err != nil {
		return err
	}

	return nil
}
