.PHONY: build clean dep
GIT_SHA=`git rev-parse --short HEAD`

build:
	@echo "Building blockchain..."
	@mkdir -p bin
	@go build -ldflags "-X main.GitSHA=${GIT_SHA}" -o bin/blockchain .

clean:
	@rm -f bin/*

dep:
	@dep ensure
	@dep prune
