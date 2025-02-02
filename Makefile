GOIMP = "gofmt"
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
BINARY_NAME=go-streamer
OUTPUT_DIR=bin

build: create-build-dir
	go build -o $(OUTPUT_DIR)/$(BINARY_NAME) cmd/main.go

run: build
	./$(OUTPUT_DIR)/$(BINARY_NAME)

fmt: ## Run gofmt for all .go files
	@$(GOIMP) -w $(GOFMT_FILES)
	@#$(GOFMT) -w $(GOFMT_FILES)

create-build-dir:
	mkdir -p $(OUTPUT_DIR)
