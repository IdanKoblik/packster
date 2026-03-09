GO := go

COVER_OUT := coverage.out
COVER_HTML := coverage.html
SERVICE_NAME := artifactor

.PHONY: all
all: build

.PHONY: build
build:
	$(GO) build -ldflags "-X 'main.BUILD_TIME=$$(date +%Y-%m-%dT%H:%M:%S)'" -o bin/$(SERVICE_NAME) ./cmd

.PHONY: test
test:
	$(GO) test -v ./...

# Unit tests only: -short skips long-running / network-heavy tests.
.PHONY: test-unit
test-unit:
	$(GO) test -short -v ./...

# Run unit tests with coverage and produce both a summary and an HTML report.
# -coverpkg=./... instruments every package in the module so that packages
# exercised transitively through dependencies also appear in the report.
.PHONY: cover
cover:
	$(GO) test -short -race -coverprofile=$(COVER_OUT) -covermode=atomic -coverpkg=./... ./...
	$(GO) tool cover -func=$(COVER_OUT)
	$(GO) tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo ""
	@echo "HTML report written to $(COVER_HTML)"

# Run unit + integration tests with merged coverage.
# Uses go tool covdata to properly merge two separate runs so both
# test types contribute without double-counting overlapping lines.
# cmd/ is excluded from coverpkg because main packages are not unit-testable.
.PHONY: cover-integration
cover-integration:
	@mkdir -p .covunit .covint .covmerged
	GOCOVERDIR=.covunit $(GO) test -short -race -cover -coverpkg=./internal/...,./pkg/... ./...
	GOCOVERDIR=.covint  $(GO) test       -race -cover -coverpkg=./internal/...,./pkg/... ./...
	$(GO) tool covdata merge -i=.covunit,.covint -o=.covmerged
	$(GO) tool covdata textfmt -i=.covmerged -o=$(COVER_OUT)
	@rm -rf .covunit .covint .covmerged
	$(GO) tool cover -func=$(COVER_OUT)
	$(GO) tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo ""
	@echo "HTML report written to $(COVER_HTML)"


.PHONY: clean
clean:
	rm -rf bin $(COVER_OUT) $(COVER_HTML) .covunit .covint .covmerged

.PHONY: run
run: build
	./bin/$(SERVICE_NAME)
