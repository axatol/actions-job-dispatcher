vet:
	go vet ./...

lint:
	helm lint ./charts/actions-job-dispatcher

deps:
	go mod download

build: build-dispatcher-binary

image: build-dispatcher-image

build-dispatcher-binary:
	go build -o ./bin/dispatcher ./cmd/dispatcher/main.go

build-dispatcher-image:
	docker build -t public.ecr.aws/axatol/actions-job-dispatcher:latest .
