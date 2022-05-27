package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	// Specify the directory for template layouts
	LayoutDir string = "views/layouts/"
	// Specifu the extension for template files
	TemplateExt string = ".gohtml"
	// Specify the views directory
	TemplateDir = "views/"
)

type View struct {
	// Template stores a parsed template file
	Template *template.Template
	Layout   string
}

// Render is used to execute a template of a View object
// The data interface is passed through to the template
func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

// ServeHTTP is used to call Handle on a view object
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := v.Render(w, nil)
	if err != nil {
		panic(err)
	}
}

// layoutFiles globs all layout templates and returns an array of template name strings
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}
	return files
}

// addTemplatePath prepends all input values with the TemplateDir
func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// addTemplateExt appends all input values with the TemplateExt
func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}

// NewView creates a new view with a base layout template and any needed files for that template
func NewView(layout string, files ...string) *View {

	addTemplatePath(files)
	addTemplateExt(files)

	// Add the footer and layout to each template file
	files = append(files, layoutFiles()...)

	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

//
