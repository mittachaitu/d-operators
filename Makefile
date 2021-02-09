# Fetch the latest tags & then set the package version
PACKAGE_VERSION ?= $(shell git fetch --all --tags | echo "" | git describe --always --tags)
ALL_SRC = $(shell find . -name "*.go" | grep -v -e "vendor")

# We are using docker hub as the default registry
IMG_NAME ?= dope
IMG_REPO ?= mayadataio/dope

all: bins

bins: vendor test $(IMG_NAME)

$(IMG_NAME): $(ALL_SRC)
	@echo "+ Generating $(IMG_NAME) binary"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
		go build -o $@ ./cmd/main.go

$(ALL_SRC): ;

# go mod download modules to local cache
# make vendored copy of dependencies
# install other go binaries for code generation
.PHONY: vendor
vendor: go.mod go.sum
	@GO111MODULE=on go mod download
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor

.PHONY: test
test: 
	@go test ./... -cover

.PHONY: testv
testv:
	@go test ./... -cover -v -args --logtostderr -v=2

.PHONY: integration-test
integration-test:
	# Uncomment to list verbose output
	# @go test ./... -cover --tags=integration -v -args --logtostderr -v=1
	@go test ./... -cover --tags=integration

.PHONY: declarative-test-suite
declarative-test-suite: 
	@cd test/declarative && ./suite.sh

.PHONY: integration-test-suite
integration-test-suite:
	@cd test/integration && ./suite.sh

.PHONY: image
image:
	docker build -t $(IMG_REPO):$(PACKAGE_VERSION) -t $(IMG_REPO):latest .

.PHONY: push
push: image
	docker push $(IMG_REPO):$(PACKAGE_VERSION)

.PHONY: push-latest
push-latest: image
	docker push $(IMG_REPO):latest
