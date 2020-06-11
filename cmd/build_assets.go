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
	var data = "package main\n\nimport \"text/template\"\n"
	for _, name := range os.Args[1:] {
		t, err := ioutil.ReadFile(name)
		if err != nil {
			panic(err)
		}

		// actually try to parse template
		_ = template.Must(template.New(name).Parse(string(t)))

		// write to file
		name := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
		data = data + "\nvar " + name + "T = template.Must(template.New(\"" + name + "\").Parse(" + fmt.Sprintf("%#v", string(t)) + "))\n"
	}
	err := ioutil.WriteFile("assets.go", []byte(data), 0644)
	if err != nil {
		panic(err)
	}
}
