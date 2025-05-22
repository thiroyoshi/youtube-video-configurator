# Blog Post - Local Execution

This directory contains the code for running the blog-post Cloud Function locally directly from the command line.

## Prerequisites

- Go 1.23.9 (for Cloud Functions compatibility)
- A properly configured `config.json` file in the working directory (same format as required by the Cloud Function)

## Running Locally

1. Navigate to this directory:
   ```
   cd cmd/blog-post
   ```

2. Build the local version:
   ```
   go build
   ```

3. Run the function directly:
   ```
   ./blog-post
   ```

   This will execute the blog post function immediately without starting a server.

4. For quick testing during development, you can use `go run`:
   ```
   go run main.go
   ```

## Configuration

Ensure that you have a `config.json` file in the working directory with the following structure:

```json
{
  "openai_api_key": "your_openai_api_key",
  "hatena_id": "your_hatena_id",
  "hatena_blog_id": "your_hatena_blog_id",
  "hatena_api_key": "your_hatena_api_key"
}
```

## Notes

- The local version accesses the same core blog post function that is deployed as a Cloud Function
- Any changes you make to the core logic in `src/blog-post/main.go` will affect both the local and cloud versions