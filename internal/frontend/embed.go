package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed all:dist
var distFS embed.FS

// Handler returns a Fiber handler that serves the embedded frontend files
func Handler() fiber.Handler {
	distSubFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	return filesystem.New(filesystem.Config{
		Root:         http.FS(distSubFS),
		Index:        "index.html",
		NotFoundFile: "index.html", // SPA fallback
		Browse:       false,
	})
}

// StaticHandler returns a handler for static assets (CSS, JS, etc.)
func StaticHandler() fiber.Handler {
	distSubFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	return filesystem.New(filesystem.Config{
		Root:   http.FS(distSubFS),
		Browse: false,
	})
}
