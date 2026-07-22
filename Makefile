SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

# -------------------------
# Color definitions
# -------------------------
COLOR_RESET := \033[0m
COLOR_BLACK := \033[0;30m
COLOR_RED := \033[0;31m
COLOR_GREEN := \033[0;32m
COLOR_YELLOW := \033[0;33m
COLOR_BLUE := \033[0;34m
COLOR_MAGENTA := \033[0;35m
COLOR_CYAN := \033[0;36m
COLOR_WHITE := \033[0;37m
COLOR_BOLD := \033[1m
COLOR_UNDERLINE := \033[4m

# Icons
ICON_OK := ✅
ICON_INFO := ℹ️
ICON_WARN := ⚠️
ICON_ERROR := ❌
ICON_RUN := 🚀
ICON_BUILD := 🔨
ICON_CLEAN := 🧹
ICON_DOCKER := 🐳
ICON_DB := 🗄️
ICON_SHELL := 💻

# -------------------------
# Configurable variables
# -------------------------
COMPOSE_FILES ?= docker-compose.yml
APP_ENV ?=
COMPOSE_BIN ?=

BUILD_DIR := build
GO_CMD := go

# -------------------------
# Helper functions
# -------------------------
define print_header
	@echo -e "$(COLOR_BOLD)$(COLOR_CYAN)$(ICON_INFO) $1$(COLOR_RESET)"
endef

define print_success
	@echo -e "$(COLOR_GREEN)$(ICON_OK) $1$(COLOR_RESET)"
endef

define print_warning
	@echo -e "$(COLOR_YELLOW)$(ICON_WARN) $1$(COLOR_RESET)"
endef

define print_error
	@echo -e "$(COLOR_RED)$(ICON_ERROR) $1$(COLOR_RESET)"
endef

define print_info
	@echo -e "$(COLOR_BLUE)$(ICON_INFO) $1$(COLOR_RESET)"
endef

define print_running
	@echo -e "$(COLOR_MAGENTA)$(ICON_RUN) $1$(COLOR_RESET)"
endef

# -------------------------
# Compose command detection
# -------------------------
ifeq ($(strip $(COMPOSE_BIN)),)
COMPOSE_BIN := $(shell command -v docker-compose 2>/dev/null || true)
ifeq ($(strip $(COMPOSE_BIN)),)
	ifneq ($(shell command -v docker 2>/dev/null || true),)
	COMPOSE_BIN := docker compose
	endif
endif
endif

ifeq ($(strip $(COMPOSE_BIN)),)
$(error "$(COLOR_RED)$(ICON_ERROR) No docker or docker-compose found in PATH.$(COLOR_RESET)")
endif

ifeq ($(APP_ENV),develop)
COMPOSE_FILES := $(COMPOSE_FILES) -f docker-compose-dev.yml
$(call print_info,Development environment detected - adding docker-compose-dev.yml)
endif

## Load .env if present
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

# -------------------------
# Helper: run docker compose
# -------------------------
define dc
$(COMPOSE_BIN) -f $(COMPOSE_FILES) $(1)
endef

# -------------------------
# Run-args handling
# -------------------------
ALL_GOALS := $(MAKECMDGOALS)
POS_ARGS := $(filter-out $(firstword $(ALL_GOALS)),$(ALL_GOALS))
SERVICE ?= $(firstword $(POS_ARGS))
CMD ?= $(strip $(wordlist 2,$(words $(POS_ARGS)),$(POS_ARGS)))

# -------------------------
# Targets
# -------------------------
.PHONY: help env init up build-up build-no-cache status down purge logs \
		redis-shell psql-shell lint test test-coverage test-integration build-clean build-arm \
		build-linux sh exec clean restart ps version health config-info seed \
		doc doc-serve

help: ## Show this help
	@echo -e "$(COLOR_BOLD)$(COLOR_CYAN)$(ICON_INFO) Available targets:$(COLOR_RESET)\n"
	@awk 'BEGIN {FS = "##"; printf "$(COLOR_BOLD)Usage:$(RESET)\n  make $(COLOR_GREEN)<target>$(COLOR_RESET) [VARIABLE=value]\n\n$(COLOR_BOLD)Targets:$(COLOR_RESET)\n"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) $(COLOR_BLUE)%s$(COLOR_RESET)\n", $$1, $$2 } END { print "" }' $(MAKEFILE_LIST)

env: ## Create .env from example if missing
	@if [ -e ./.env ]; then \
		echo -e "$(COLOR_YELLOW)$(ICON_WARN) .env already exists$(COLOR_RESET)"; \
	else \
		cp -v ./.env.example ./.env; \
		echo -e "$(COLOR_GREEN)$(ICON_OK) Created .env from .env.example$(COLOR_RESET)"; \
	fi

up: ## Create and start containers (detached)
	$(call print_running,"Starting containers...")
	$(call dc,up -d)
	$(call print_success,"Containers started successfully")

build-up: ## Build images and start containers (detached)
	$(call print_running,"Building and starting containers...")
	$(call dc,up --build -d)
	$(call print_success,"Containers built and started")

build-no-cache: ## Build images without cache
	$(call print_header,"Building images without cache...")
	$(call dc,build --no-cache)
	$(call print_success,"Images built without cache")

status: ## Show currently running containers
	$(call print_info,"Container status:")
	$(call dc,ps $(SERVICE))

down: ## Stop containers (keeps volumes)
	$(call print_warning,"Stopping containers...")
	$(call dc,down --remove-orphans $(RUN_ARGS))
	$(call print_success,"Containers stopped")

purge: ## Stop containers and remove volumes
	$(call print_warning,"Stopping containers and removing volumes...")
	$(call dc,down --remove-orphans --volumes $(RUN_ARGS))
	$(call print_success,"Containers stopped and volumes removed")

restart: down up ## Restart containers
	$(call print_success,"Containers restarted")

logs: ## Tail logs for a service (use SERVICE= or positional argument)
	$(call print_info,"Tailing logs for '$(SERVICE)'")
	$(call dc,logs -f $(SERVICE))

exec: ## Run CMD inside SERVICE. Provide SERVICE and CMD.
ifndef SERVICE
	$(call print_error,"SERVICE is required. e.g. make exec SERVICE=app CMD='bash'")
	@exit 1
endif
ifndef CMD
	$(call print_error,"CMD is required. e.g. make exec SERVICE=app CMD='bash'")
	@exit 1
endif
	$(call print_running,"Executing command in $(SERVICE): $(CMD)")
	$(call dc,exec $(SERVICE) sh -lc '$(CMD)')

redis-shell: ## Open redis-cli inside redis container
	$(call print_running,"Opening Redis CLI...")
	$(call dc,exec redis redis-cli)

psql-shell: ## Open psql shell inside postgres container
	$(call print_running,"Opening PostgreSQL shell...")
	$(call dc,exec postgres psql -U $(DB_USER) -d $(DB_NAME))

lint: ## Run golangci-lint
	$(call print_header,"Running golangci-lint...")
	golangci-lint run
	$(call print_success,"Linting completed")

test: ## Run all tests
	$(call print_header,"Running tests...")
	go test ./... -short -v
	$(call print_success,"Tests passed")

test-coverage: ## Run tests with coverage report
	$(call print_header,"Running tests with coverage...")
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	$(call print_success,"Coverage report: coverage.html")

test-integration: ## Run integration tests (requires running postgres)
	$(call print_header,"Running integration tests...")
	go test ./internal/repository/postgres/... ./pkg/database/... -v
	$(call print_success,"Integration tests passed")

build-clean: ## Remove build directory
	$(call print_warning,"Cleaning build directory...")
	@rm -rf $(BUILD_DIR)
	$(call print_success,"Build directory cleaned")

build-arm: build-clean ## Build macOS/arm64 binary
	$(call print_header,"Building darwin/arm64 binary...")
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -v -a -ldflags "-w -s" -o $(BUILD_DIR)/task-manager .
	$(call print_success,"darwin/arm64 build completed")

build-linux: build-clean ## Build linux/amd64 binary
	$(call print_header,"Building linux/amd64 binary...")
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -v -a -ldflags "-w -s" -o $(BUILD_DIR)/task-manager .
	$(call print_success,"linux/amd64 build completed")

clean: ## Cleanup build artifacts
	$(call print_header,"Cleaning up...")
	@rm -rf $(BUILD_DIR) coverage.out coverage.html
	@echo -e "$(COLOR_GREEN)$(ICON_CLEAN) Cleanup completed$(COLOR_RESET)"

version: ## Show version information
	$(call print_header,"Version Information:")
	@echo -e "$(COLOR_BOLD)Docker:$(COLOR_RESET) $$(docker --version 2>/dev/null || echo 'Not installed')"
	@echo -e "$(COLOR_BOLD)Docker Compose:$(COLOR_RESET) $$(docker-compose --version 2>/dev/null || echo 'Using docker compose')"
	@echo -e "$(COLOR_BOLD)Go:$(COLOR_RESET) $$(go version 2>/dev/null || echo 'Not installed')"

health: ## Check service health
	$(call print_header,"Checking service health...")
	$(call dc,ps --filter status=running)
	@echo -e "\n$(COLOR_BOLD)Health check completed$(COLOR_RESET)"

config-info: ## Show current configuration
	$(call print_header,"Current Configuration:")
	@echo -e "$(COLOR_BOLD)COMPOSE_BIN:$(COLOR_RESET) $(COMPOSE_BIN)"
	@echo -e "$(COLOR_BOLD)COMPOSE_FILES:$(COLOR_RESET) $(COMPOSE_FILES)"
	@echo -e "$(COLOR_BOLD)APP_ENV:$(COLOR_RESET) $(if $(APP_ENV),$(APP_ENV),default)"
	@echo -e "$(COLOR_BOLD)BUILD_DIR:$(COLOR_RESET) $(BUILD_DIR)"

seed: ## Seed database with test data
	$(call print_header,"Seeding database...")
	psql -h localhost -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f seed/tasks.sql
	$(call print_success,"Database seeded successfully")

doc: ## Show go doc for all packages
	$(call print_header,"Go documentation...")
	@for pkg in $$(go list ./... 2>/dev/null); do \
		echo -e "\n$(COLOR_BOLD)$$pkg$(COLOR_RESET)"; \
		go doc -all "$$pkg" 2>/dev/null || echo "  (no doc)"; \
	done

doc-serve: ## Start local godoc server on :6060
	$(call print_running,"Starting godoc server on http://localhost:6060")
	@godoc -http=:6060
