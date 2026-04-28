GO := go
NPM := npm
NAME := packster
WEB_DIR := web
UI_OUT := internal/ui/static

COVER_OUT := coverage.out
COVER_HTML := coverage.html

all: build

ui:
	cd $(WEB_DIR) && $(NPM) install && $(NPM) run build

build: ui
	$(GO) build -ldflags "-X 'main.BUILD_TIME=$$(date +%Y-%m-%dT%H:%M:%S)'" -o bin/$(NAME) ./cmd

test:
	$(GO) test -v ./...

cover-integration:
	$(GO) test -race -coverprofile=$(COVER_OUT) -coverpkg=./internal/... ./internal/...
	$(GO) tool cover -func=$(COVER_OUT)
	$(GO) tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo ""
	@echo "HTML report written to $(COVER_HTML)"

clean:
	rm -rf bin $(COVER_OUT) $(COVER_HTML) .covunit .covint .covmerged $(UI_OUT)/assets $(UI_OUT)/index.html

run: build
	./bin/$(NAME)

.PHONY: all build ui test cover-integration clean run
