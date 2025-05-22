package main

import (
	"fmt"
	"os"

	blogpost "thiroyoshi.com/blog-post"
)

// Main function for local execution
func main() {
	fmt.Println("Executing blog post function...")

	// Call the exported function directly
	err := blogpost.BlogPost()
	if err != nil {
		fmt.Printf("Error executing blog post: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Blog post execution completed successfully.")
}
