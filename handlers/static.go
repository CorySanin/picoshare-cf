package handlers

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"
)

//go:embed static
var staticFS embed.FS

// Because we embed static files in the Go binary, they cannot change while the
// process is running. Use the process start time and request path to generate a
// stable ETag for each resource without hashing every resource at startup.
var staticETagSeed = time.Now().String()

// serveStaticResource serves any static file under the ./handlers/static
// directory.
func serveStaticResource() http.HandlerFunc {
	fSys, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	server := http.FileServer(http.FS(fSys))

	return func(w http.ResponseWriter, r *http.Request) {
		etag := sha256.Sum256([]byte(staticETagSeed + r.URL.Path))
		w.Header().Set("ETag", fmt.Sprintf(`"%x"`, etag))
		w.Header().Set("Cache-Control", staticCacheControl)

		// FileServer honors conditional and range requests after we set the ETag.
		server.ServeHTTP(w, r)
	}
}
