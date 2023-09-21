package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	workDir, _ := os.Getwd()
	staticFiles := http.Dir(filepath.Join(workDir, "assets"))

	router := chi.NewRouter()
	
	router.Use(middleware.Logger)

	//Public Routes
	router.Group(func(r chi.Router) {
		router.Get("/", indexHandler)
	})

	//API Routes
	router.Group(func(r chi.Router) {
		//Middleware
		r.Use(ValidateHTMXRequest)
		//Subroutes
		r.Route("/api", func(r chi.Router) {
			r.Post("/signup-newsletter", postHandler)
		})
	})


	FileServer(router, "/assets", staticFiles)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	err := server.ListenAndServe()

	if err != nil {
		fmt.Println("failed to listen to server", err)
	}
}

func ValidateHTMXRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	template := template.Must(template.ParseFiles("index.html"))
	template.Execute(w, nil)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")

	log.Println("email", email)

	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
