.PHONY: build clean dep

build:
	@echo "Building blockchain..."
	@mkdir -p bin
	@go build -o bin/blockchain .

clean:
	@rm -f bin/*

dep:
	@dep ensure
	@dep prune
