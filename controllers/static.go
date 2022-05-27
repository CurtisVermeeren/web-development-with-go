package controllers

import "github.com/curtisvermeeren/web-development-with-go/views"

// Static controller is used for views that have just one action
type Static struct {
	Home    *views.View
	Contact *views.View
	Faq     *views.View
}

// NewStatic creates views for all the static pages
func NewStatic() *Static {
	return &Static{
		Home:    views.NewView("bootstrap", "static/home"),
		Contact: views.NewView("bootstrap", "static/contact"),
		Faq:     views.NewView("bootstrap", "static/faq"),
	}
}
