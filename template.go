package goutils

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig"
)

// Loader loads layout, partial and template files from a specified
// directory.
//
// Dealing with partials and layouts in template files is notoriously tricky
// using the standard Golang template parser. This Loader makes is
// easier to simply load a template by name and parse one of the blocks within
// that template file.
type Loader struct {
	Dir         string
	LayoutsDir  string
	PartialsDir string
	Suffix      string
}

// NewTemplateLoader creates a Loader with the recommended layouts
// directory "/layouts" and partials directory "/partials". These directories
// are relative from the specified template directory.
func NewTemplateLoader(dir string) *Loader {
	return &Loader{
		Dir:         dir,
		LayoutsDir:  "/layouts",
		PartialsDir: "/partials",
		Suffix:      ".tmpl",
	}
}

// LoadAllTemplates creates separate templates for each template
// in the template directory.
func (t *Loader) LoadAllTemplates() (map[string]*template.Template, error) {
	files, err := os.ReadDir(t.Dir)
	if err != nil {
		return nil, fmt.Errorf("reading dir %q: %w", t.Dir, err)
	}

	tmpls := make(map[string]*template.Template)

	for _, f := range files {
		tmplName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))

		info, err := f.Info()
		if err != nil {
			return nil, fmt.Errorf("reading file info of %q: %w", f.Name(), err)
		}

		if t.isTemplate(info) {
			tmpl, err := t.LoadTemplate(tmplName)
			if err != nil {
				return nil, fmt.Errorf("loading template file %q: %w", tmplName, err)
			}

			tmpls[tmplName] = tmpl
		}
	}

	return tmpls, nil
}

// LoadTemplate loads a single template file and any partials and layout templates
// from the template directory.
func (t *Loader) LoadTemplate(templateName string) (*template.Template, error) {
	fs, err := t.getTemplateFileNames(templateName)
	if err != nil {
		return nil, err
	}

	return template.Must(template.New(templateName).Funcs(sprig.FuncMap()).ParseFiles(fs...)), nil
}

// GetTemplateFileNames returns all template filenames that should be loaded,
// including the layout and partial templates (located in ./layouts and ./partials).
func (t *Loader) getTemplateFileNames(templateName string) ([]string, error) {
	templatePath := t.Dir + "/" + templateName + t.Suffix
	if _, err := os.Stat(templatePath); err != nil {
		return nil, fmt.Errorf("reading template file info of %q: %w", templatePath, err)
	}

	fs := make([]string, 0)
	fs = append(fs, templatePath)

	// Load partials
	filenames, err := t.dirFilenames(t.Dir + t.PartialsDir)
	if err != nil {
		return nil, err
	}

	fs = append(fs, filenames...)

	// Load layouts
	layouts, err := t.dirFilenames(t.Dir + t.LayoutsDir)
	if err != nil {
		return nil, err
	}

	fs = append(fs, layouts...)

	return fs, nil
}

func (t *Loader) dirFilenames(dir string) ([]string, error) {
	fs := make([]string, 0)

	partials, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading dir %q: %w", dir, err)
	}

	for _, f := range partials {
		info, err := f.Info()
		if err != nil {
			return nil, fmt.Errorf("reading file info of %q: %w", f.Name(), err)
		}

		if t.isTemplate(info) {
			fs = append(fs, dir+"/"+f.Name())
		}
	}

	return fs, nil
}

func (t *Loader) isTemplate(f os.FileInfo) bool {
	return !f.IsDir() && len(f.Name()) > 0 && strings.HasSuffix(f.Name(), t.Suffix)
}
