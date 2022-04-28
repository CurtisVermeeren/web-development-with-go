package views

import (
	"html/template"
	"log"
)

type View struct {
	// Template stores a parsed template file
	Template *template.Template
}

func NewView(files ...string) *View {
	// Add the footer to each template file
	files = append(files, "views/layouts/footer.gohtml")
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}

	return &View{
		Template: t,
	}
}
