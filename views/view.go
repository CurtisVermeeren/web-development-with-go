package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"

	"github.com/curtisvermeeren/web-development-with-go/context"
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
func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var vd Data
	// Ensure the data sent to the template is wrapped in the Data struct
	switch d := data.(type) {
	case Data:
		vd = d
	default:
		vd = Data{
			Yield: data,
		}
	}

	vd.User = context.User(r.Context())
	fmt.Println(vd.User)
	// Attempt to write the template to a buffer instead of directly to ResponseWriter
	// This will prevent status 200 being written in the Response before all errors are checked
	var buf bytes.Buffer
	err := v.Template.ExecuteTemplate(&buf, v.Layout, vd)
	if err != nil {
		http.Error(w, "Something went wrong. If the problem persists, please contact support.", http.StatusInternalServerError)
	}
	// Copy the buffer to the ResponseWriter if no errors occur
	io.Copy(w, &buf)
}

// ServeHTTP is used to call Handle on a view object
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
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
