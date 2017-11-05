.PHONY: build clean

build:
	@echo "Building blockchain..."
	@mkdir -p bin
	@go build -o bin/blockchain .

clean:
	@rm -f bin/*