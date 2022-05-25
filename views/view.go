package views

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var (
	// Specify the directory for template layouts
	LayoutDir string = "views/layouts/"
	// Specift the extension for template files
	TemplateExt string = ".gohtml"
)

type View struct {
	// Template stores a parsed template file
	Template *template.Template
	Layout   string
}

// Render is used to execute a template of a View object
// The data interface is passed through to the template
func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

// layout files globs all templates and returns an array of template name strings
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func NewView(layout string, files ...string) *View {
	// Add the footer and layout to each template file
	files = append(files, layoutFiles()...)

	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}
