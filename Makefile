########################################################
#                       Makefile                       #
########################################################

# Default target
all: format build

########################################################
#                         Setup                        #
########################################################

# Generate versioning information
TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
VERSION := $(TAG:v%=%)
ifneq ($(COMMIT), $(TAG_COMMIT))
    VERSION := $(VERSION)-next-$(COMMIT)-$(DATE)
endif
ifneq ($(shell git status --porcelain),)
    VERSION := $(VERSION)-dirty
endif


########################################################
#                       Building                       #
########################################################

# Target for building the application in all directories
build:; GOEXPERIMENT=rangefunc go build ./...

# Run the example applications
run-%:; go run ./examples/$*/main.go start

# Format
lint: |
	GOEXPERIMENT=rangefunc go run github.com/golangci/golangci-lint/cmd/golangci-lint run

# Test
# TODO - Remove the GOEXPERIMENT=rangefunc once Go 1.23 is released
test: |
	GOEXPERIMENT=rangefunc go test -v ./...

# Format
format: |
	GOEXPERIMENT=rangefunc go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix

generate: |
	GOEXPERIMENT=rangefunc go generate ./...

tidy: |
	go mod tidy