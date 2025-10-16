ARTIFACT_NAME := $(shell grep "^module " go.mod | awk '{print $$2}' | xargs basename)
SERVICE_ACCOUNT_KEY_FILE := ./tools/serviceaccount/zitadel-admin-sa.json

##@ Default target
default: fmt lint install generate

##@ Build Targets
build:
	go build -v ./...

artifact:
	go build -gcflags="all=-N -l" -o $(ARTIFACT_NAME)

install: build
	go install -v ./...

##@ Code Quality
lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

##@ Testing
test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	@if [ -z "$$ZITACTL_SERVICE_ACCOUNT_KEY" ] && [ -f $(SERVICE_ACCOUNT_KEY_FILE) ]; then \
		ZITACTL_SERVICE_ACCOUNT_KEY=$$(cat $(SERVICE_ACCOUNT_KEY_FILE)); \
	fi; \
	if [ -z "$$ZITACTL_SERVICE_ACCOUNT_KEY" ]; then \
		echo "Error: ZITACTL_SERVICE_ACCOUNT_KEY is not set and $(SERVICE_ACCOUNT_KEY_FILE) does not exist" >&2; \
		exit 1; \
	fi; \
	TF_ACC=1 \
	ZITACTL_DOMAIN=$${ZITACTL_DOMAIN:-localhost} \
	ZITACTL_SKIP_TLS_VERIFICATION=$${ZITACTL_SKIP_TLS_VERIFICATION:-true} \
	ZITACTL_SERVICE_ACCOUNT_KEY="$$ZITACTL_SERVICE_ACCOUNT_KEY" \
	go test -v -cover -timeout 120m ./...

test-release:
	# This target needs the GPG_FINGERPRINT environment variable to be set.
	# Use e.g. `gpg --list-secret-keys --with-colons | awk -F: '/^fpr:/ {print $10; exit}'` to acquire it.
	@source ./get-gpg-passphrase.sh && \
	goreleaser release --snapshot --clean

##@ Test infrastructure
zitadel-up:
	docker compose -f ./tools/docker-compose.yml up -d --wait

zitadel-down:
	docker compose -f ./tools/docker-compose.yml down -v

zitadel-logs:
	docker compose -f ./tools/docker-compose.yml logs -f

##@ Help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@sed -n -e 's/^##@ \(.*\)/\n\1:/p' -e 's/^\([a-zA-Z_-]*\):.*/  \1/p' $(MAKEFILE_LIST)

.PHONY: default build artifact install lint generate fmt test testacc test-release zitadel-down zitadel-logs help