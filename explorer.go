package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// html template
const pageTemplate = `
<DOCTYPE html>
<html>
	<head>
	<title>Remote file explorer</title>
	<style>
	a {
		text-decoration: none;
	}
	body {
		font-family: verdana;
		font-size: 14px;
		padding-left: 10px"
	}
	</style>
	</head>
	<body>
		{{range .Items}}
		{{if .IsDir}}
		<div class="directory">
		<a href='{{ .Url }}'>{{ .Name }}</a>
		</div>
		{{else}}
		<div class="file">		
		<span>{{ .Name }}</span>
		</div>
		{{end}}
		{{else}}
		<div>
		<span>No items found.</span>
		</div>
		{{end}}
	</body>
</html>
`

// struct to hold file/folder data
type ItemViewModel struct {
	Name  string
	Url   string
	IsDir bool
}

// struct to hold page data
type PageViewModel struct {
	Items []ItemViewModel
}

// generate and write template html
func writeTemplate(w http.ResponseWriter, src string, path string, files []os.FileInfo) {
	pageViewModel := PageViewModel{}

	// add up directory item
	prevPath := trimSuffix(filepath.Dir(path), ".")

	if prevPath != path {
		prevDirHref := fmt.Sprintf("/browse/%s", convertPathToURL(prevPath))
		prevDirItem := ItemViewModel{Name: ".. Go Up", IsDir: true, Url: prevDirHref}
		pageViewModel.Items = append(pageViewModel.Items, prevDirItem)
	}

	// add current directory items
	for _, f := range files {
		href := fmt.Sprintf("%s/%s", src, url.QueryEscape(f.Name()))
		item := ItemViewModel{Name: f.Name(), IsDir: f.IsDir(), Url: href}
		pageViewModel.Items = append(pageViewModel.Items, item)
	}

	// execute html template
	t, err := template.New("page").Parse(pageTemplate)
	if err == nil {
		err = t.Execute(w, pageViewModel)
	}

	// write error if template generation fails
	if err != nil {
		writeError(w, "Internal Error")
	}
}

// remove suffix in a string
func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}

	return s
}

// convert filepath to a url
func convertPathToURL(src string) string {
	path := src

	// if on windows machine remove colon after the drive letter
	// and convert backward slashes to forward slashes
	if runtime.GOOS == "windows" {
		path = strings.Replace(path, ":", "", 1)
		path = strings.Replace(path, "\\", "/", -1)
	}

	return url.QueryEscape(path)
}

// convert a given url to file path
func convertUrlToPath(src string) (string, error) {
	path := src

	// remove "browse" prefix
	path = strings.Replace(path, "/browse", "", 1)

	// unescape query string
	path, _ = url.QueryUnescape(path)

	// remove trailing slash
	trimSuffix(path, "/")

	// if on windows machine fix drive letters with colon
	// and convert forward slashes to backward slashes
	if runtime.GOOS == "windows" {
		drives := []string{"c", "d", "e", "f"}
		for _, d := range drives {
			if strings.HasPrefix(path, "/"+d) {
				path = strings.Replace(path, "/"+d, d+":", 1)
			}
		}

		path = strings.Replace(path, "/", "\\", -1)
	}

	return path, nil
}

// write error message
func writeError(w http.ResponseWriter, message string) {
	log.Print(message)
	fmt.Fprintf(w, "%s", message)
}

// hanlde /browse route
func browseHandler(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Path

	path, err := convertUrlToPath(src)

	if err != nil {
		writeError(w, "Invalid Path")
	} else {
		files, err := ioutil.ReadDir(path)

		if err != nil {
			writeError(w, "Not Found : "+path)
		} else {
			w.Header().Set("Content-Type", "text/html")
			writeTemplate(w, src, path, files)
		}
	}
}

func main() {
	// map routes
	http.HandleFunc("/browse/", browseHandler)

	// start http server
	log.Fatal(http.ListenAndServe(":3000", nil))
}
