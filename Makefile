DOCKER_IMAGE_NAME = public.ecr.aws/axatol/actions-job-dispatcher

GO_BUILD_LDFLAGS = -X 'main.buildCommit=$(shell git rev-parse HEAD)'
GO_BUILD_LDFLAGS += -X 'main.buildTime=$(shell date +"%Y-%m-%dT%H:%M:%S%z")'

vet:
	go vet ./...

lint:
	helm lint ./charts/actions-job-dispatcher

deps:
	go mod download

build: build-dispatcher-binary

image: build-dispatcher-image

build-dispatcher-binary:
	go build -o ./bin/dispatcher -ldflags="$(GO_BUILD_LDFLAGS)" ./cmd/dispatcher/main.go

build-dispatcher-image:
	docker build -t $(DOCKER_IMAGE_NAME):latest .
