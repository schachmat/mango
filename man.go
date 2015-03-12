package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	validPage *regexp.Regexp
)

func fail(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	io.WriteString(w, message)
}

func stripHeader(r io.ReadCloser, w io.Writer) {
	inHeader := true
	for s := bufio.NewScanner(r); s.Scan(); {
		if !inHeader {
			io.WriteString(w, s.Text()+"\n")
		}
		if s.Text() == "" {
			inHeader = false
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if !validPage.MatchString(r.RequestURI) {
		fail(w, http.StatusNotFound, "could not match request to a manpage")
		return
	}

	page := validPage.FindStringSubmatch(r.RequestURI)[1:]
	man := exec.Command("man", "--path", page[1])
	if page[0] != "" {
		man = exec.Command("man", "--path", page[0], page[1])
	}

	fname, err := man.Output()
	if err != nil {
		fail(w, http.StatusNotFound, "could not get filename for man"+page[0]+" "+page[1]+":"+err.Error())
		return
	}

	bzcat := exec.Command("bzcat", strings.TrimSpace(string(fname)))
	man2html := exec.Command("man2html", "-p", "-M", "", "-H", "localhost:8626")

	i, o := io.Pipe()
	bzcat.Stdout = o
	bzcat.Stderr = os.Stderr
	man2html.Stdin = i
	fr, fw := io.Pipe()
	man2html.Stdout = fw

	err = bzcat.Start()
	if err != nil {
		log.Fatal("bzcat.Start")
		fail(w, http.StatusInternalServerError, "Could not start bzcat: "+err.Error())
		bzcat.Wait()
		return
	}

	err = man2html.Start()
	if err != nil {
		log.Fatal("man2html.Start")
		fail(w, http.StatusInternalServerError, "Could not start man2html: "+err.Error())
		bzcat.Wait()
		o.Close()
		man2html.Wait()
		return
	}

	go stripHeader(fr, w)

	err = bzcat.Wait()
	if err != nil {
		fail(w, http.StatusInternalServerError, "Could not wait for bzcat: "+err.Error())
		o.Close()
		man2html.Wait()
		return
	}

	o.Close()

	err = man2html.Wait()
	if err != nil {
		fail(w, http.StatusInternalServerError, "Could not wait for man2html: "+err.Error())
		return
	}
}

func init() {
	validPage = regexp.MustCompile(`/(?:([0-8n]p?)\+)?(.+)`)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe("localhost:8626", nil)
}
