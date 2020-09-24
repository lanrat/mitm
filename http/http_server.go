package main

import (
	"flag"
	"fmt"
	"log"
	"mitm/http/server"
	"net/http"
	"os"
	"path"
	"strings"
)

var (
	listenAddr = flag.String("listen", "", "ip to listen on")
	base       = "data"
)

func main() {
	flag.Parse()
	// get server and start application
	httpServer, err := server.New(*listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	Start(httpServer)
	log.Printf("Server starting on %s", *listenAddr)
	log.Fatal(httpServer.Start())
}

// object to hold application context and persistent storage
type appContext struct {
	//templates *template.Template
	//session   sessions.Store
}

// Start entry point for starting application
// adds routes to the server so that the correct handlers are registered
func Start(server *server.Server) {
	var app appContext
	// compile all templates and cache them
	//app.templates = template.Must(template.New("main").ParseGlob("templates/*.tmpl"))
	//app.session = server.Session

	// set the routes
	//server.Get("/test", app.TestHandler)
	server.GetPrefix("/", app.DomainHandler)

}

func (app *appContext) TestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World\n")
}

func (app *appContext) DomainHandler(w http.ResponseWriter, r *http.Request) {
	host := strings.ToLower(r.Host)
	urlPath := r.URL.Path
	filePath := path.Join(base, host, urlPath)

	if !fileExists(filePath) {
		// 404
		http.NotFoundHandler().ServeHTTP(w, r)
		return
	}
	http.ServeFile(w, r, filePath)

}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	//log.Printf("checking file: %s", filename)
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()

}
