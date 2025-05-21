# Blog Post - Local Execution

This directory contains the code for running the blog-post Cloud Function locally.

## Prerequisites

- Go 1.24 or later
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

3. Run the local server:
   ```
   ./blog-post
   ```

   The server will start on port 8080 by default. You can change the port by setting the PORT environment variable:
   ```
   PORT=9000 ./blog-post
   ```

4. Access the server:
   - Open a web browser and navigate to `http://localhost:8080`
   - Or use curl: `curl http://localhost:8080`

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

- The local version accesses the same `BlogPost` function that is deployed as a Cloud Function
- Any changes you make to the core logic in `src/blog-post/main.go` will affect both the local and cloud versions