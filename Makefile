all: lint docker

lint:
	golangci-lint run --timeout 5m ./... -v

package:
	CGO_ENABLED=0 go build -ldflags="-extldflags '-static'" -o k8s-multi-secret-to-file .

docker:
	docker build -t k8s-multi-secret-to-file:local .