cd functions/video-converter
gcloud functions deploy video-converter --gen2 --runtime=go116 --region=asia-northeast1 --entry-point=VideoConverter --trigger-http --allow-unauthenticated