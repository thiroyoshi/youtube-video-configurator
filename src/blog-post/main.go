package blogpost

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

// blogPost is an HTTP Cloud Function.
func blogPost(w http.ResponseWriter, r *http.Request) {
	err := RunBlogPost()
	if err != nil {
		slog.Error("Error executing blog post", "error", err)
		return
	}

	_, err = fmt.Fprint(w, "Blog post successfully executed")
	if err != nil {
		slog.Error("Error writing response", "error", err)
	}
}

// BlogPost is the exported version of blogPost for external use
func BlogPost(w http.ResponseWriter, r *http.Request) {
	blogPost(w, r)
}

func init() {
	// Register HTTP handler for Cloud Functions
	functions.HTTP("BlogPost", blogPost)
}
