default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc-up: | ssh-keys
	docker compose up -d
	./scripts/create-bastion-account.sh

testacc-down:
	docker compose down --volumes

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

ssh-keys:
	./scripts/create-ssh-keys.sh

clean: testacc-down
	@rm -rf ssh-keys/

.PHONY: fmt lint test testacc build install generate testacc
