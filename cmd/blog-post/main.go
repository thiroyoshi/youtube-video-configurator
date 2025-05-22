package main

import (
	"fmt"
	"os"

	blogpost "thiroyoshi.com/blog-post"
)

// Main function for local execution
func main() {
	fmt.Println("Executing blog post function...")

	// Call the function directly without starting a server
	err := blogpost.RunBlogPost()
	if err != nil {
		fmt.Printf("Error executing blog post: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Blog post execution completed successfully.")
}
