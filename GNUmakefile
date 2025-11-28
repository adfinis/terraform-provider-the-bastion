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

local-install:
	go install -v ./...
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/adfinis/bastion/0.0.1/linux_amd64
	mv ~/go/bin/terraform-provider-bastion ~/.terraform.d/plugins/registry.terraform.io/adfinis/bastion/0.0.1/linux_amd64

.PHONY: fmt lint test testacc build install generate testacc
