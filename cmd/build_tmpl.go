package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	var data = `package main

import (
	"errors"
	"io"
	"text/template"

	"github.com/pitr/gig"
)

type Template struct{}

var ErrUnknownTemplate = errors.New("unknown template")

func (*Template) Render(w io.Writer, name string, data interface{}, c gig.Context) error {
	var t *template.Template
	switch name {
`
	for _, name := range os.Args[1:] {
		t, err := ioutil.ReadFile(name)
		if err != nil {
			panic(err)
		}

		// actually try to parse template
		_ = template.Must(template.New(name).Parse(string(t)))

		// write to file
		name := strings.TrimSuffix(strings.TrimPrefix(name, "tmpl/"), filepath.Ext(name))
		data += "\tcase \"" + name + "\":\n\t\tt = template.Must(template.New(\"" + name + "\").Parse(" + fmt.Sprintf("%#v", string(t)) + "))\n"
	}

	data += `	default:
		return ErrUnknownTemplate
	}
	return t.ExecuteTemplate(w, name, data)
}
`
	err := ioutil.WriteFile("tmpl.go", []byte(data), 0644)
	if err != nil {
		panic(err)
	}
}
