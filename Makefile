# convert-starter のデプロイ
.PHONY: deploy-convert-starter

deploy-convert-starter:
	cd src/convert-starter && \
	zip -r ../../artifacts/convert-starter.zip . && \
	gsutil cp ../../artifacts/convert-starter.zip gs://video-converter-src-bucket/convert-starter.zip && \
	cd ../../infra && terraform apply -auto-approve -target=module.convert_starter

# video-converter のデプロイ
.PHONY: deploy-video-converter

deploy-video-converter:
	cd src/video-converter && \
	zip -r ../../artifacts/video-converter.zip . && \
	gsutil cp ../../artifacts/video-converter.zip gs://video-converter-src-bucket/video-converter.zip && \
	cd ../../infra && terraform apply -auto-approve -target=module.video_converter

# blog-post のデプロイ
.PHONY: deploy-blog-post

deploy-blog-post:
	cd src/blog-post && \
	zip -r ../../artifacts/blog-post_1234.zip . && \
	gsutil cp ../../artifacts/blog-post_1234.zip gs://video-converter-src-bucket/blog-post_1234.zip && \
	cd ../../infra && terraform apply -auto-approve -target=module.blog_post -var="short_sha=1234"
