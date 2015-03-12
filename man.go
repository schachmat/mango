package main

import (
	"bufio"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

var (
	listenAddr string
	validPage *regexp.Regexp
)

func fail(w http.ResponseWriter, status int, err error) {
	log.Println(err)
	w.WriteHeader(status)
	io.WriteString(w, err.Error())
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
		fail(w, http.StatusNotFound, errors.New(r.RequestURI+" is no valid manual"))
		return
	}

	page := validPage.FindStringSubmatch(r.RequestURI)[1:]
	man := exec.Command("man", "--path", page[1])
	if page[0] != "" {
		man = exec.Command("man", "--path", page[0], page[1])
		page[1] = strings.Join(page, " ")
	}

	fname, err := man.Output()
	if err != nil {
		fail(w, http.StatusNotFound, errors.New("manual for "+page[1]+" not found"))
		return
	}

	bzcat := exec.Command("bzcat", strings.TrimSpace(string(fname)))
	man2html := exec.Command("man2html", "-p", "-M", "", "-H", listenAddr)

	// Setup pipeline: bzcat -> man2html -> strip http header -> ResponseWriter
	p1r, p1w := io.Pipe()
	p2r, p2w := io.Pipe()
	bzcat.Stdout = p1w
	man2html.Stdin = p1r
	man2html.Stdout = p2w
	go stripHeader(p2r, w)

	err = bzcat.Start()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		bzcat.Wait()
		return
	}

	err = man2html.Start()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		p1w.Close()
		bzcat.Wait()
		man2html.Wait()
		return
	}

	go stripHeader(p2r, w)

	err = bzcat.Wait()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		p1w.Close()
		man2html.Wait()
		return
	}

	p1w.Close()

	err = man2html.Wait()
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
}

func init() {
	validPage = regexp.MustCompile(`/(?:([0-8n]p?)\+)?(.+)`)
	flag.StringVar(&listenAddr, "listen", "localhost:8626", "On which address and port should we listen? Default is localhost:8626")
}

func main() {
	flag.Parse()
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
