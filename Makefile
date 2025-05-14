# convert-starter のデプロイ
.PHONY: deploy-convert-starter

deploy-convert-starter:
	cd src/convert-starter && \
	zip -r ../../artifacts/convert-starter.zip . && \
	gsutil cp ../../artifacts/convert-starter.zip gs://video-converter-src-bucket/convert-starter.zip && \
	cd ../../infra && terraform apply -auto-approve

# video-converter のデプロイ
.PHONY: deploy-video-converter

deploy-video-converter:
	cd src/video-converter && \
	zip -r ../../artifacts/video-converter.zip . && \
	gsutil cp ../../artifacts/video-converter.zip gs://video-converter-src-bucket/video-converter.zip && \
	cd ../../infra && terraform apply -auto-approve
