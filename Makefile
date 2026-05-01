GO ?= go
GOTOOLCHAIN ?= $(shell $(GO) env GOVERSION)
GO_TEST = GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) test
COVER_PROFILE ?= coverage.out
COVER_HTML ?= coverage.html

.PHONY: test test-integration coverage

test:
	$(GO_TEST) ./...

test-integration:
	$(GO_TEST) -tags=integration ./...

coverage:
	$(GO_TEST) -coverprofile=$(COVER_PROFILE) ./...
	$(GO) tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)
