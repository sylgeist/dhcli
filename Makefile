.PHONY: help
.DEFAULT_GOAL := all
VERSION = "$(shell git rev-parse HEAD)"

help: ## Display this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: all
all: format build test ## Reformat, build, and test

.PHONY: build
build: bin/dhcli ## Build binary

.PHONY: clean
clean: ## Clean build output
	rm -rf bin/

.PHONY: bin/dhcli
bin/dhcli: ## Build the 'bin/dhcli' binary
	CGO_ENABLED=0 go build -ldflags "-X do/doge/version.commit=$(VERSION)" -o bin/ ./cmd/dhcli/...

.PHONY: test
test: ## Run unit tests
	go test -coverprofile=untagged.coverage -v ./...

.PHONY: all.coverage
all.coverage: ## Generate combined coverage report
	echo "mode: set" > all.coverage && \
		cat *.coverage | grep -v 'mode: ' | sort | uniq >> all.coverage

.PHONY: view-coverage
view-coverage: all.coverage ## View coverage report in-browser
	go tool cover -html=all.coverage

.PHONY: format
format: ## Reformat code
	go fmt ./...

.PHONY: version
version: .version ## Display current version
	@cat .version

.PHONY: .version
.version:
	@echo $(VERSION) > .version

.PHONY: update-go-deps
update-go-deps: ## Update Go vendor dependencies
	@echo ">> updating Go dependencies"
	@for m in $$(go list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		go get $$m; \
	done
	go mod tidy
ifneq (,$(wildcard vendor))
	go mod vendor
endif
