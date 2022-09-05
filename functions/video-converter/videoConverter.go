// Package helloworld provides a set of Cloud Functions samples.
package videoConverter

import (
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("VideoConverter", videoConverter)
}

// HelloGet is an HTTP Cloud Function.
func videoConverter(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}
