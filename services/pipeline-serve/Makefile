run: build
	go run .

build: build-fetch-image build-grayify-image build-pipeline-serve
	docker-compose build
	docker run --privileged -p=8080:8080 --link=redis:redis --link=s3:s3 -v=${PWD}/pipeline.yml:/go/src/pipeline.yml -v=/var/run/docker.sock:/var/run/docker.sock process-some-image:0.1.0

build-fetch-image:
	cd ../fetch-image && GOOS=linux go build -o fetch-image

build-grayify-image:
	cd ../grayify-image && GOOS=linux go build -o grayify-image

build-pipeline-serve:
	GOOS=linux go build -o pipeline-serve

goofys:
	docker run -it --rm --privileged -e BUCKET="tmp" -e AWS_ACCESS_KEY_ID=minioadmin -e AWS_SECRET_ACCESS_KEY=minioadmin cloudposse/goofys