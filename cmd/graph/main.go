package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/n3wscott/knap/pkg/config"
	"github.com/n3wscott/knap/pkg/graph"
	"html/template"
	"image"
	"image/jpeg"
	"io/ioutil"
	"k8s.io/client-go/dynamic"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	cluster    string
	kubeconfig string
)

type envConfig struct {
	// Namespace this pod exists in.
	Namespace string `envconfig:"POD_NAMESPACE" required:"true"`
}

func init() {
	flag.StringVar(&cluster, "cluster", "",
		"Provide the cluster to test against. Defaults to the current cluster in kubeconfig.")

	var defaultKubeconfig string
	if usr, err := user.Current(); err == nil {
		defaultKubeconfig = path.Join(usr.HomeDir, ".kube/config")
	}

	flag.StringVar(&kubeconfig, "kubeconfig", defaultKubeconfig,
		"Provide the path to the `kubeconfig` file.")
}

var client dynamic.Interface
var env envConfig

func main() {
	flag.Parse()

	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	cfg, err := config.BuildClientConfig(kubeconfig, cluster)
	if err != nil {
		log.Fatalf("Error building kubeconfig", err)
	}

	client = dynamic.NewForConfigOrDie(cfg)

	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/", handler)

	log.Println("Listening on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func favicon(w http.ResponseWriter, r *http.Request) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, img, nil); err != nil {
		log.Println("unable to encode image.")
	}

	writeBytes(w, buffer.Bytes(), "jpg")
}

func getQueryParam(r *http.Request, key string) string {
	keys, ok := r.URL.Query()[key]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

var defaultPage = "html"     // or img
var defaultFormat = "svg"    // or png
var defaultFocus = "trigger" // or png

func handler(w http.ResponseWriter, r *http.Request) {
	page := getQueryParam(r, "page")
	if page == "" {
		page = defaultPage
	}

	format := getQueryParam(r, "format")
	if format == "" {
		format = defaultFormat
	}

	focus := getQueryParam(r, "focus")
	if focus == "" {
		focus = defaultFocus
	}

	var dotGraph string

	switch focus {
	case "sub", "subs", "subscription", "subscriptions":
		dotGraph = graph.ForSubscriptions(client, env.Namespace)
	case "broker", "trigger", "triggers":
		fallthrough
	default:
		dotGraph = graph.ForTriggers(client, env.Namespace)
	}

	file, err := dotToImage(format, []byte(dotGraph))
	if err != nil {
		log.Printf("dotToImage error %s", err)
		return
	}
	img, err := ioutil.ReadFile(file)

	defer os.Remove(file) // clean up

	if page == "html" {
		writeBytesWithTemplate(w, img, format)
	} else {
		writeBytes(w, img, format)
	}

}

var dot string

func dotToImage(format string, b []byte) (string, error) {

	if dot == "" {
		var err error
		dot, err = exec.LookPath("dot")
		if err != nil {
			log.Fatalln("unable to find program 'dot', please install it or check your PATH")
		}
	}

	var img = filepath.Join(os.TempDir(), fmt.Sprintf("graph.%s", format))

	cmd := exec.Command(dot, fmt.Sprintf("-T%s", format), "-o", img) //, tmpfile.Name())
	cmd.Stdin = bytes.NewBuffer(b)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return img, nil
}

var Template = `<!DOCTYPE html>
<html lang="en"><head></head>
<body><img src="data:{{.Format}},{{.Image}}"></body>`

func writeBytesWithTemplate(w http.ResponseWriter, b []byte, format string) {
	if format == "svg" {
		_, _ = w.Write([]byte(`<!DOCTYPE html><html lang="en"><head></head><body>`))
		_, _ = w.Write(b)
		_, _ = w.Write([]byte(`</body></html>`))
		return
	}

	data := map[string]interface{}{
		"Image":  base64.StdEncoding.EncodeToString(b),
		"Format": fmt.Sprintf("image/%s;base64", format),
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
