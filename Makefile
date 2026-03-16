.PHONY: build test install clean

build:
	go build -o terraform-provider-sops-sakura-kms .

test:
	go test -v ./...

install:
	go install .

clean:
	rm -f terraform-provider-sops-sakura-kms
