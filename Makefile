APP_NAME=rates
DOCKER_IMAGE=rates-image

build:
	go build -o rates cmd/main.go


docker-build:
	docker build -t $(DOCKER_IMAGE) .

test:
	go test ./... -v


run:
	go run cmd/main.go

lint:
	golangci-lint run
