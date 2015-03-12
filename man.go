package main

import (
	"io"
	"fmt"
	"regexp"
	"net/http"
	"os/exec"
)

var (
	validPage *regexp.Regexp
)

func handler(w http.ResponseWriter, r *http.Request) {
	if !validPage.MatchString(r.RequestURI) {
		return
	}
	io.WriteString(w, fmt.Sprintln(validPage.FindStringSubmatch(r.RequestURI)))

	page := validPage.FindStringSubmatch(r.RequestURI)[1:]

	var man = exec.Command("man", "--path", page[1])
	if page[0] != "" {
		man = exec.Command("man", "--path", page[0], page[1])
	}
	fname, err := man.Output()
	if err != nil {
		io.WriteString(w, "failed")
		return
	}
	io.WriteString(w, string(fname) + "\n")

	//TODO: bzcat manpage | man2html -p
}

func init() {
	validPage = regexp.MustCompile(`/(?:([0-8n]p?)\+)?(.+)`)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8626", nil)
}
