# Hatena Blog Post Generator

This Cloud Function automatically generates and posts blog content to Hatena Blog using RSS feeds and OpenAI.

## Local Execution

You can run this function locally using the command-line tool in the `cmd/blog-post` directory:

```bash
# Navigate to the cmd directory
cd ../../cmd/blog-post

# Build the executable
go build

# Run the function directly
./blog-post
```

The function will execute immediately without starting a server, making it easy to run on demand.

## Configuration

Make sure you have a `config.json` file in the working directory with the following structure:

```json
{
  "openai_api_key": "your-openai-api-key",
  "hatena_id": "your-hatena-id",
  "hatena_blog_id": "your-hatena-blog-id",
  "hatena_api_key": "your-hatena-api-key"
}
```

## Cloud Deployment

For deployment to Google Cloud Functions, use the following command:

```bash
gcloud functions deploy blog-post \
  --runtime go124 \
  --trigger-http \
  --allow-unauthenticated
```