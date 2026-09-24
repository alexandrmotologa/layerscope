.PHONY: all build test clean ui run demo

all: build

ui:
	cd ui && npm install && npm run build

build: ui
	go build -tags embed_ui -o bin/layerscope.exe ./cmd/layerscope

build-cli:
	go build -o bin/layerscope.exe ./cmd/layerscope

test:
	go test -v ./...

demo: build
	./bin/layerscope.exe analyze --demo

clean:
	rm -rf bin/ pkg/server/dist/ ui/dist/
