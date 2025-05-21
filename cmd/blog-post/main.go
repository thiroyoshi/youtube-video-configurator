package main

import (
	"fmt"
	"net/http"
	"os"
	
	blogpost "thiroyoshi.com/blog-post"
)

// Main function for local execution
func main() {
	// Create a simple HTTP handler that calls the blog post function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Executing blog post function locally...")
		
		// Call the Cloud Function directly
		blogpost.BlogPost(w, r)
		
		fmt.Println("Blog post execution completed.")
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting local server on port %s...\n", port)
	fmt.Printf("Open http://localhost:%s in your browser or use curl to execute the blog post function\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}