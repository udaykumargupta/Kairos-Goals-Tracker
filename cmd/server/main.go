// Command server runs Kairos as an ordinary long-lived HTTP server: the same API
// the serverless entrypoint exposes, plus the static frontend. Useful for local
// development and for hosts that run a container rather than functions.
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/udaykumargupta/kairos/kairos"
)

func main() {
	cfg, err := kairos.LoadConfig()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}
	app, err := kairos.New(cfg)
	if err != nil {
		log.Fatalf("startup error: %v", err)
	}

	api := app.Routes()
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "frontend"
	}
	files := http.FileServer(http.Dir(staticDir))

	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The API owns /api/** and the Android asset-links file; everything else is
		// the static app. There is no SPA fallback — deep links are query params on "/".
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/.well-known/assetlinks.json" {
			api.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}
		files.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Kairos listening on http://localhost:%s (static from %q)", port, staticDir)
	if err := http.ListenAndServe(":"+port, root); err != nil {
		log.Fatal(err)
	}
}
