GOLANGCI_VERSION := 0.0.38
IMAGE_TAG := $(if $(IMAGE_TAG),$(IMAGE_TAG),$(shell git rev-parse --short HEAD))

.PHONY: docker
docker:
	docker build -t ${IMAGE_TAG} -f Dockerfile .

.PHONY: dep
dep:
	go mod download
	go mod tidy

.PHONY: build
build: dep
	go build -o ./mittwald-container-action cmd/action/main.go

.PHONY: lint
lint: dep
	docker run --rm -t \
    		-v $(shell go env GOPATH):/go \
    		-v ${CURDIR}:/app \
    		-v $(HOME)/.cache:/home/mittwald-golangci/.cache \
    		-w /app \
    		-e GOFLAGS="-buildvcs=false" \
    		-e GOLANGCI_ADDITIONAL_YML="/app/build/ci/.golangci.yml" \
    		quay.io/mittwald/golangci-lint:$(GOLANGCI_VERSION) \
        		golangci-lint run -v --fix ./...

.PHONY: test
test:
	go test -v -failfast -count=1 ./cmd/...