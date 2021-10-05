package beebot

import (
	"embed"
	"io/fs"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

//go:embed templates/*.html
var embeddedFS embed.FS
var files fs.FS

// ServeWeb configures and starts the webserver
func (b *BeeBot) ServeWeb() {
	router := chi.NewRouter()

	files = getFileSystem(b.debug)

	api := b.apiEndpoints()
	router.Mount("/api/v1", api)

	router.HandleFunc("/", staticPage("index.html"))
	router.HandleFunc("/filters", staticPage("filters.html"))
	router.HandleFunc("/config", staticPage("config.html"))
	router.HandleFunc("/log", staticPage("log.html"))

	b.nav["Filters"] = "/filters"
	b.nav["Config"] = "/config"
	b.nav["Log"] = "/log"

	// Don't want to block for this (later)
	baseAddr := b.c.Get("baseaddr", DefaultAddr)
	log.Fatal().
		Err(http.ListenAndServe(baseAddr, router)).
		Msg("HTTP server")
}

func staticPage(templatePath string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		index, err := fs.ReadFile(files, templatePath)
		if err != nil {
			log.Error().Err(err).Msg("Could not read template")
		}
		w.Write(index)
	}
}

func getFileSystem(useOS bool) fs.FS {
	if useOS {
		log.Print("using live mode")
		return os.DirFS("templates")
	}

	log.Print("using embed mode")
	fsys, err := fs.Sub(embeddedFS, "templates")
	if err != nil {
		log.Error().Err(err).Msg("Could not load file templates")
	}

	return fsys
}
