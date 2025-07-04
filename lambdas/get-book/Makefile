.PHONY: all build zip clean fmt vet tidy check

# Allow overriding GOOS and GOARCH. Defaults to linux/amd64 for Lambda.
GOOS ?= linux
GOARCH ?= amd64

TARGET_NAME=get-book
BINARY_NAME=$(TARGET_NAME)-$(GOOS)-$(GOARCH)

all: build

build:
	@mkdir -p bin
	@echo "Building $(BINARY_NAME)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/$(BINARY_NAME) main.go
	@echo "Successfully built bin/$(BINARY_NAME)"

# Create a zip archive for Lambda deployment (always linux/amd64, binary is 'bootstrap').
zip:
	@mkdir -p bin dist
	@echo "Building for Lambda and creating zip archive..."
	GOOS=linux GOARCH=amd64 go build -o bin/bootstrap main.go
	cd bin && zip ../dist/$(TARGET_NAME).zip bootstrap
	@echo "Successfully created dist/$(TARGET_NAME).zip"

check: fmt vet

fmt:
	@echo "Formatting code..."
	gofmt -w .

vet:
	@echo "Vetting code..."
	go vet ./...

tidy:
	@echo "Tidying dependencies..."
	go mod tidy

clean:
	@echo "Cleaning up..."
	rm -rf bin dist
