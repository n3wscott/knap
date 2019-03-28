package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func main() {
	http.HandleFunc("/", handler)
	log.Println("Listening on 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func getQueryParam(r *http.Request, key string) string {
	keys, ok := r.URL.Query()[key]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

var defaultFormat = "png"

func handler(w http.ResponseWriter, r *http.Request) {

	page := getQueryParam(r, "page")

	format := getQueryParam(r, "format")

	if format == "" {
		format = defaultFormat
	}

	//var img image.Image

	// /user/bin/dot

	graph := `
graph {
    a -- b;
    b -- c;
    a -- c;
    d -- c;
    e -- c;
    e -- a;
}
  `

	file, err := dotToImage(format, []byte(graph))
	if err != nil {
		log.Printf("dotToImage error %s", err)
		return
	}
	log.Printf("dotToImage image %s", file)
	img, err := ioutil.ReadFile(file)

	if page == "html" {
		writeBytesWithTemplate(w, img, format)
	} else {
		writeBytes(w, img, format)
	}
}

var dotExe string

func dotToImage(format string, dot []byte) (string, error) {
	if dotExe == "" {
		dot, err := exec.LookPath("dot")
		if err != nil {
			log.Fatalln("unable to find program 'dot', please install it or check your PATH")
		}
		dotExe = dot
	}

	var img = filepath.Join(os.TempDir(), fmt.Sprintf("graph.%s", format))

	cmd := exec.Command(dotExe, fmt.Sprintf("-T%s", format), "-o", img)
	cmd.Stdin = bytes.NewReader(dot)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return img, nil
}

var Template = `<!DOCTYPE html>
<html lang="en"><head></head>
<body><img src="data:{{.Format}},{{.Image}}"></body>`

func writeBytesWithTemplate(w http.ResponseWriter, b []byte, format string) {
	var data map[string]interface{}

	if format == "svg" {
		data = map[string]interface{}{
			"Image":  string(b),
			"Format": fmt.Sprintf("image/%s+xml;utf8", format),
		}
	} else {
		data = map[string]interface{}{
			"Image":  base64.StdEncoding.EncodeToString(b),
			"Format": fmt.Sprintf("image/%s;base64", format),
		}
	}
	if tmpl, err := template.New("image").Parse(Template); err != nil {
		log.Println("unable to parse image template.")
	} else {
		if err = tmpl.Execute(w, data); err != nil {
			log.Println("unable to execute template.")
		}
	}
}

// writeImage encodes an image 'img' in jpeg format and writes it into ResponseWriter.
func writeBytes(w http.ResponseWriter, b []byte, format string) {
	w.Header().Set("Content-Type", "image/"+format)
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	if _, err := w.Write(b); err != nil {
		log.Println("unable to write image.")
	}
}
