GO := go
NAME := packster

COVER_OUT := coverage.out
COVER_HTML := coverage.html

all: build

build:
	$(GO) build -ldflags "-X 'main.BUILD_TIME=$$(date +%Y-%m-%dT%H:%M:%S)'" -o bin/$(NAME) ./cmd

test:
	$(GO) test -v ./...

clean:
	rm -rf bin $(COVER_OUT) $(COVER_HTML) .covunit .covint .covmerged

run: build
	./bin/$(NAME)

.PHONY: all, build, clean, run
