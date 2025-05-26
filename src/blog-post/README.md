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
  "hatena_api_key": "your-hatena-api-key",
  "slack_webhook_url": "your-slack-webhook-url"
}
```

## Cloud Deployment

For deployment to Google Cloud Functions, use the following command:

```bash
gcloud functions deploy blog-post \
  --runtime go124 \
  --trigger-http \
  --set-env-vars HATENA_ID=your-hatena-id,HATENA_BLOG_ID=your-hatena-blog-id \
  --set-secrets OPENAI_API_KEY=blog-post-openai-api-key:latest,HATENA_API_KEY=blog-post-hatena-api-key:latest,SLACK_WEBHOOK_URL=blog-post-slack-webhook-url:latest \
  --allow-unauthenticated
```

## Slack Notifications

The function sends a notification to Slack when a new blog post is published. To set up Slack notifications:

1. Create a new Slack app in your workspace: https://api.slack.com/apps
2. Enable Incoming Webhooks for your app
3. Create a new webhook URL for the channel where you want to receive notifications
4. Set the webhook URL as the `SLACK_WEBHOOK_URL` environment variable or in your `config.json` file

### Troubleshooting Slack Notifications

If you encounter issues with Slack notifications (e.g., 403 Forbidden errors):

1. Verify your webhook URL is correct and hasn't been revoked
2. Make sure the webhook URL is properly formatted and begins with `https://hooks.slack.com/services/`
3. Check that your Slack app has the necessary permissions for the channel
4. Look for detailed error messages in the Cloud Function logs

The notification system will automatically retry up to 3 times with increasing delays between attempts in case of temporary failures.