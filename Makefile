.PHONY: build dev generate-bindings test vet fmt clean

build:
	wails3 build

dev:
	wails3 dev

generate-bindings:
	wails3 generate bindings

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin
