#
# MIT License
#
# (C) Copyright 2021-2025 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#

# ──────────────────────────────────────────────────────────────────────────────
#  Project settings
# ──────────────────────────────────────────────────────────────────────────────

NAME   := cani
MODULE := github.com/Cray-HPE/$(NAME)
SHELL  := /bin/bash -o pipefail

lc = $(subst A,a,$(subst B,b,$(subst C,c,$(subst D,d,$(subst E,e,$(subst F,f,$(subst G,g,$(subst H,h,$(subst I,i,$(subst J,j,$(subst K,k,$(subst L,l,$(subst M,m,$(subst N,n,$(subst O,o,$(subst P,p,$(subst Q,q,$(subst R,r,$(subst S,s,$(subst T,t,$(subst U,u,$(subst V,v,$(subst W,w,$(subst X,x,$(subst Y,y,$(subst Z,z,$1))))))))))))))))))))))))))

# ──────────────────────────────────────────────────────────────────────────────
#  Colors  (disable with NO_COLOR=1)
# ──────────────────────────────────────────────────────────────────────────────

ifndef NO_COLOR
  CLR_RESET  := \033[0m
  CLR_BOLD   := \033[1m
  CLR_GREEN  := \033[32m
  CLR_YELLOW := \033[33m
  CLR_CYAN   := \033[36m
  CLR_RED    := \033[31m
else
  CLR_RESET  :=
  CLR_BOLD   :=
  CLR_GREEN  :=
  CLR_YELLOW :=
  CLR_CYAN   :=
  CLR_RED    :=
endif

INFO  = @printf "$(CLR_CYAN)>>> %s$(CLR_RESET)\n"
OK    = @printf "$(CLR_GREEN) ✔  %s$(CLR_RESET)\n"
WARN  = @printf "$(CLR_YELLOW) ⚠  %s$(CLR_RESET)\n"
ERR   = @printf "$(CLR_RED) ✖  %s$(CLR_RESET)\n"

# ──────────────────────────────────────────────────────────────────────────────
#  Platform detection
# ──────────────────────────────────────────────────────────────────────────────

ifeq ($(NAME),)
export NAME := $(shell basename $(shell pwd))
endif

ifeq ($(ARCH),)
export ARCH := $(shell uname -m)
endif

ifeq ($(VERSION),)
export VERSION := $(shell git describe --tags 2>/dev/null | tr -s '-' '~' | sed 's/^v//')
endif

ifeq ($(GOOS),)
OS := $(shell uname)
export GOOS := $(call lc,$(OS))
endif

ifeq ($(GOARCH),)
	ifeq "$(ARCH)" "aarch64"
		export GOARCH=arm64
	else ifeq "$(ARCH)" "arm64"
		export GOARCH=arm64
	else ifeq "$(ARCH)" "x86_64"
		export GOARCH=amd64
	endif
endif

# ──────────────────────────────────────────────────────────────────────────────
#  Git / build metadata
# ──────────────────────────────────────────────────────────────────────────────

GIT_COMMIT  := $(shell git rev-parse --short HEAD)
GIT_BRANCH  := $(shell git rev-parse --abbrev-ref HEAD)
GIT_VERSION := $(shell git describe --tags 2>/dev/null || echo "$(GIT_COMMIT)")
BUILDTIME   := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

CHANGELOG_VERSION := $(shell grep -m1 ' \[[0-9]*.[0-9]*.[0-9]*\]' CHANGELOG.MD 2>/dev/null | sed -e "s/\].*$$//" -e "s/^.*\[//")

# ──────────────────────────────────────────────────────────────────────────────
#  Linker flags
# ──────────────────────────────────────────────────────────────────────────────

git_dirty := $(shell git status -s)
go_ldflags := -s -w
ifeq ($(git_dirty),)
	go_ldflags += -X $(MODULE)/cmd.GitTreeState=clean
else
	go_ldflags += -X $(MODULE)/cmd.GitTreeState=dirty
endif
go_ldflags += -X $(MODULE)/cmd.GitTag=$(VERSION)
go_ldflags += -X $(MODULE)/cmd.BuildDate=$(BUILDTIME)

# ──────────────────────────────────────────────────────────────────────────────
#  Paths
# ──────────────────────────────────────────────────────────────────────────────

BIN_DIR         := $(CURDIR)/bin
BUILD_DIR       ?= $(CURDIR)/dist/rpmbuild
TEST_OUTPUT_DIR ?= $(CURDIR)/build/results
SPEC_FILE       ?= $(NAME).spec
SOURCE_NAME     ?= $(NAME)-$(VERSION)
SOURCE_PATH     := $(BUILD_DIR)/SOURCES/$(SOURCE_NAME).tar.bz2

# ──────────────────────────────────────────────────────────────────────────────
#  Default target
# ──────────────────────────────────────────────────────────────────────────────

.DEFAULT_GOAL := all

.PHONY: all
all: fmt vet bin ## Build the project (fmt → vet → compile)
	$(OK) "all targets complete"

# ──────────────────────────────────────────────────────────────────────────────
#  Build
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: bin
bin: ## Compile the binary into bin/
	$(INFO) "building $(NAME) ($(GOOS)/$(GOARCH))"
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(NAME) -ldflags '$(go_ldflags)'
	$(OK) "$(BIN_DIR)/$(NAME)"

.PHONY: install
install: ## go install the binary
	$(INFO) "installing $(NAME)"
	go install -ldflags '$(go_ldflags)'
	$(OK) "installed"

.PHONY: clean
clean: ## Remove build artifacts
	$(INFO) "cleaning"
	go clean -i ./...
	rm -rf $(BIN_DIR) $(BUILD_DIR)
	rm -f $(CURDIR)/build/results/coverage/* $(CURDIR)/build/results/unittest/*
	$(OK) "clean"

# ──────────────────────────────────────────────────────────────────────────────
#  Code quality
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: fmt
fmt: ## Format Go source files
	$(INFO) "formatting"
	@go fmt ./...
	$(OK) "formatted"

.PHONY: vet
vet: ## Run go vet
	$(INFO) "vetting"
	@go vet ./...
	$(OK) "vetted"

.PHONY: lint
lint: ## Run static analysis
	$(INFO) "linting"
	golint -set_exit_status ./cmd/...
	golint -set_exit_status ./internal/...
	golint -set_exit_status ./pkg/...
	$(OK) "linted"

# ──────────────────────────────────────────────────────────────────────────────
#  Testing
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: test
# Run all tests unit + functional + integration (edge disabled-see below)
test: utest ftest itest # etest
	$(OK) "all tests passed"

.PHONY: utest
utest: bin ## Run unit tests
	$(INFO) "running unit tests"
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -cover \
	    $(MODULE)/internal/config \
	    $(MODULE)/pkg/datastores \
	    $(MODULE)/pkg/devicetypes \
	    $(MODULE)/pkg/provider/csm/client \
			$(MODULE)/pkg/provider/csm/import \
  		$(MODULE)/pkg/provider/csm/transform \
	    $(MODULE)/pkg/provider/example/export \
	    $(MODULE)/pkg/provider/example/import \
	    $(MODULE)/pkg/provider/example/transform \
	    $(MODULE)/pkg/provider/nautobot/export \
	    $(MODULE)/pkg/provider/ochami/transform \
	    $(MODULE)/pkg/provider/redfish/import \
	    $(MODULE)/pkg/provider/redfish/transform \
	    $(MODULE)/pkg/visual
	$(OK) "unit tests passed"

.PHONY: ftest
ftest: bin ## Run functional tests
	$(INFO) "running functional tests"
	./spec/support/bin/cani_integrate.sh functional
	$(OK) "functional tests passed"

.PHONY: itest
itest: bin ## Run integration tests
	$(INFO) "running integration tests"
	./spec/support/bin/cani_integrate.sh integration
	$(OK) "integration tests passed"

.PHONY: etest
# disabled in make test due to EOL CSM and slow nature, but can be run independently with `make etest`
etest: bin ## Run edge-case tests
	$(INFO) "running edge tests"
	SKIP_EXTERNAL_TESTS=1 ./spec/support/bin/cani_integrate.sh edge
	$(OK) "edge tests passed"

# ──────────────────────────────────────────────────────────────────────────────
#  Dependencies
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: tidy
tidy: ## Tidy and vendor modules
	$(INFO) "tidying modules"
	go mod tidy
	go mod vendor
	$(OK) "modules tidied"

.PHONY: tools
tools: ## Install code-generation tools
	$(INFO) "installing tools"
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	$(OK) "tools installed"

# ──────────────────────────────────────────────────────────────────────────────
#  Code generation
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: nautobot_client
nautobot_client: ## Regenerate the Nautobot API client
	$(INFO) "generating Nautobot client"
	curl -sS https://raw.githubusercontent.com/nautobot/go-nautobot/refs/heads/main/api/openapi.yaml > pkg/nautobot/openapi.yml
	oapi-codegen -package nautobot -generate client,models,std-http -o pkg/nautobot/nautobot_api.go ./pkg/nautobot/openapi.yml
	$(OK) "Nautobot client generated"

.PHONY: generate-swagger-sls-client
generate-swagger-sls-client: bin/swagger-codegen-cli.jar ## Generate SLS client
	$(INFO) "generating SLS client"
	java -jar bin/swagger-codegen-cli.jar generate -i ./pkg/sls-client/openapi.yaml -l go -o ./pkg/sls-client/ -DpackageName=sls_client -t ./pkg/sls-client/templates
	go fmt ./pkg/sls-client/...
	goimports -w ./pkg/sls-client
	$(OK) "SLS client generated"

.PHONY: generate-swagger-hsm-client
generate-swagger-hsm-client: bin/swagger-codegen-cli.jar ## Generate HSM client
	$(INFO) "generating HSM client"
	java -jar bin/swagger-codegen-cli.jar generate -i ./pkg/hsm-client/openapi.yaml -l go -o ./pkg/hsm-client/ -DpackageName=hsm_client
	go fmt ./pkg/hsm-client/...
	goimports -w ./pkg/hsm-client
	$(OK) "HSM client generated"

.PHONY: generate-swagger-hpcm-client
generate-swagger-hpcm-client: bin/swagger-codegen-cli.jar ## Generate HPCM client
	$(INFO) "generating HPCM client"
	java -jar bin/swagger-codegen-cli.jar generate -i ./pkg/hpcm-client/openapi.yml -l go -o ./pkg/hpcm-client/ -DpackageName=hpcm_client
	go fmt ./pkg/hpcm-client/...
	goimports -w ./pkg/hpcm-client

# ──────────────────────────────────────────────────────────────────────────────
#  CSM
# ──────────────────────────────────────────────────────────────────────────────


CSM_CERTS_DIR ?= testdata/fixtures/csm/simulator/nginx/certs

.PHONY: csm-certs
csm-certs: ## Generate self-signed TLS certs for the CSM API gateway
	@if [ -f $(CSM_CERTS_DIR)/cert.crt ] && [ -f $(CSM_CERTS_DIR)/cert.key ]; then \
		echo "  certs already exist at $(CSM_CERTS_DIR), skipping"; \
	else \
		mkdir -p $(CSM_CERTS_DIR); \
		openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
			-subj "/C=US/ST=Minnesota/L=Bloomington/O=HPE/OU=Engineering/CN=hpe.com" \
			-out $(CSM_CERTS_DIR)/cert.crt \
			-keyout $(CSM_CERTS_DIR)/cert.key 2>/dev/null; \
		printf "$(CLR_GREEN) ✔  %s$(CLR_RESET)\n" "generated TLS certs at $(CSM_CERTS_DIR)"; \
	fi

SLS_FILE ?= testdata/fixtures/csm/sls/valid_hardware_networks.json

.PHONY: csm-up
csm-up: csm-certs ## Run the CSM simulator
	$(INFO) "starting CSM simulator (use SLS_FILE=$(SLS_FILE) to specify SLS fixture)"
	docker compose -f testdata/fixtures/csm/simulator/docker-compose.yml up -d
	$(INFO) "waiting for api-gateway to become healthy"
	@until curl -kso /dev/null https://localhost:8443/apis/sls/v1/health 2>/dev/null; do sleep 2; done
	$(INFO) "loading SLS data from $(SLS_FILE)"
	curl -k -X POST -F "sls_dump=@$(SLS_FILE)" https://localhost:8443/apis/sls/v1/loadstate -i
	$(OK) "CSM simulator running with SLS data loaded"

.PHONY: csm-down
csm-down: ## Run the CSM simulator
	$(INFO) "stopping CSM simulator"
	docker compose -f testdata/fixtures/csm/simulator/docker-compose.yml down
	$(OK) "CSM simulator stopped"

.PHONY: spec-setup
spec-setup: ## Install shellspec for BDD tests
	$(INFO) "setting up shellspec"
	@if ! [ -d ./shellspec ]; then \
	  git clone https://github.com/shellspec/shellspec.git; \
	  ln -s "$(shell pwd)"/shellspec/shellspec /usr/local/bin/; \
	fi
	$(OK) "shellspec ready"

.PHONY: load-sls
load-sls:
	$(INFO) "loading SLS data for simulator tests"
	spec/support/bin/setup_simulator.sh ./hms-simulation-environment ../testdata/fixtures/sls/no_hardware.json
	$(OK) "SLS data loaded"

# ──────────────────────────────────────────────────────────────────────────────
#  Documentation
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: venv
venv: ## Create a Python virtualenv for docs
	$(INFO) "creating virtualenv"
	virtualenv -p python3 venv
	source venv/bin/activate
	pip install -r requirements.txt
	$(OK) "virtualenv ready"

.PHONY: serve
serve: venv ## Serve docs locally with mkdocs
	$(INFO) "serving docs"
	mkdocs serve

# ──────────────────────────────────────────────────────────────────────────────
#  RPM packaging
# ────────────────────────────────────────────

rpm_prepare:
	$(INFO) "preparing RPM workspace"
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/SPECS $(BUILD_DIR)/SOURCES
	cp $(SPEC_FILE) $(BUILD_DIR)/SPECS/

.PHONY: rpm_package_source
rpm_package_source:
	$(INFO) "packaging source tarball"
	tar --transform 'flags=r;s,^,/$(SOURCE_NAME)/,' --exclude .git --exclude dist -cvjf $(SOURCE_PATH) .

.PHONY: rpm_build_source
rpm_build_source:
	$(INFO) "building source RPM"
	rpmbuild --nodeps --target $(ARCH) -ts $(SOURCE_PATH) --define "_topdir $(BUILD_DIR)"

.PHONY: rpm_build
rpm_build:
	$(INFO) "building RPM"
	rpmbuild --nodeps --target $(ARCH) -ba $(SPEC_FILE) --define "_topdir $(BUILD_DIR)"

# ──────────────────────────────────────────────────────────────────────────────
#  Misc
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: license
license: ## Run the license checker
	docker run -it --rm -v $(PWD):/github/workspace artifactory.algol60.net/csm-docker/stable/license-checker .github/workflows/ cmd/ internal pkg/hardwaretypes pkg/xname spec/ --fix

.PHONY: env
env: ## Print Go environment
	@go env

.PHONY: version
version: ## Print Go version
	@go version

.PHONY: info
info: ## Show build variables
	@printf "$(CLR_BOLD)name$(CLR_RESET)     %s\n" "$(NAME)"
	@printf "$(CLR_BOLD)version$(CLR_RESET)  %s\n" "$(VERSION)"
	@printf "$(CLR_BOLD)commit$(CLR_RESET)   %s\n" "$(GIT_COMMIT)"
	@printf "$(CLR_BOLD)branch$(CLR_RESET)   %s\n" "$(GIT_BRANCH)"
	@printf "$(CLR_BOLD)os/arch$(CLR_RESET)  %s/%s\n" "$(GOOS)" "$(GOARCH)"
	@printf "$(CLR_BOLD)dirty$(CLR_RESET)    %s\n" "$(if $(git_dirty),yes,no)"

# ──────────────────────────────────────────────────────────────────────────────
#  Help
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help
	@printf "\n$(CLR_BOLD)Usage:$(CLR_RESET)  make $(CLR_CYAN)<target>$(CLR_RESET)\n\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CLR_CYAN)%-28s$(CLR_RESET) %s\n", $$1, $$2}'
	@printf "\n"
