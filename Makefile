#
# MIT License
#
# (C) Copyright 2021-2024 Hewlett Packard Enterprise Development LP
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
NAME := cani
SHELL := /bin/bash -o pipefail
lc =$(subst A,a,$(subst B,b,$(subst C,c,$(subst D,d,$(subst E,e,$(subst F,f,$(subst G,g,$(subst H,h,$(subst I,i,$(subst J,j,$(subst K,k,$(subst L,l,$(subst M,m,$(subst N,n,$(subst O,o,$(subst P,p,$(subst Q,q,$(subst R,r,$(subst S,s,$(subst T,t,$(subst U,u,$(subst V,v,$(subst W,w,$(subst X,x,$(subst Y,y,$(subst Z,z,$1))))))))))))))))))))))))))

ifeq ($(NAME),)
export NAME := $(shell basename $(shell pwd))
endif

ifeq ($(ARCH),)
export ARCH := $(shell uname -m)
endif

ifeq ($(VERSION),)
export VERSION := $(shell git describe --tags | tr -s '-' '~' | sed 's/^v//')
endif

# By default, if these are not set then set them to match the host.
ifeq ($(GOOS),)
OS := $(shell uname)
export GOOS := $(call lc,$(OS))
endif
ifeq ($(GOARCH),)
	ifeq "$(ARCH)" "aarch64"
		export GOARCH=arm64
	else ifeq "$(ARCH)" "x86_64"
		export GOARCH=amd64
	endif
endif

GO_FILES?=$$(find . -name '*.go' |grep -v vendor)
TAG?=latest

.GIT_COMMIT=$(shell git rev-parse --short HEAD)
.GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
.GIT_COMMIT_AND_BRANCH=$(.GIT_COMMIT)-$(subst /,-,$(.GIT_BRANCH))
.GIT_VERSION=$(shell git describe --tags 2>/dev/null || echo "$(.GIT_COMMIT)")
.BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
CHANGELOG_VERSION_ORIG=$(grep -m1 \## CHANGELOG.MD | sed -e "s/\].*\$//" | sed -e "s/^.*\[//")
CHANGELOG_VERSION=$(shell grep -m1 \ \[[0-9]*.[0-9]*.[0-9]*\] CHANGELOG.MD | sed -e "s/\].*$$//" | sed -e "s/^.*\[//")
BUILD_DIR ?= $(PWD)/dist/rpmbuild
SPEC_FILE ?= ${NAME}.spec
SOURCE_NAME ?= ${NAME}-${VERSION}
SOURCE_PATH := ${BUILD_DIR}/SOURCES/${SOURCE_NAME}.tar.bz2
TEST_OUTPUT_DIR ?= $(CURDIR)/build/results

# If there are uncommitted changes, append "-dirty"
git_dirty := $(shell git status -s)
go_ldflags := -s -w
ifeq ($(git_dirty),)
	go_ldflags += -X github.com/Cray-HPE/${NAME}/cmd.GitTreeState='clean'
else
	go_ldflags += -X github.com/Cray-HPE/${NAME}/cmd.GitTreeState='dirty'
endif
go_ldflags += -X github.com/Cray-HPE/${NAME}/cmd.GitTag=$(VERSION)
go_ldflags += -X github.com/Cray-HPE/${NAME}/cmd.BuildDate=${.BUILDTIME}

.PHONY: \
	bin \
	help \
	clean \
	tools \
	test \
	vet \
	lint \
	fmt \
	tidy \
	env \
	build \
	rpm \
	doc \
	version \
	spec \
	validate-hardware-type-schemas \
	generate \
	generate-go \
	generate-swagger \
	license \
	venv

all: bin

rpm: rpm_prepare rpm_package_source rpm_build_source rpm_build

help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help               Show this help screen.'
	@echo '    clean              Remove binaries, artifacts and releases.'
	@echo '    tools              Install tools needed by the project.'
	@echo '    test               Run unit tests.'
	@echo '    vet                Run go vet.'
	@echo '    lint               Run golint.'
	@echo '    fmt                Run go fmt.'
	@echo '    tidy               Run go mod tidy.'
	@echo '    env                Display Go environment.'
	@echo '    build              Build project for current platform.'
	@echo '    rpm                Build a YUM/SUSE RPM.'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    version            Display Go version.'
	@echo ''

clean:
	go clean -i ./...
	rm -vf \
	  $(CURDIR)/build/results/coverage/* \
	  $(CURDIR)/build/results/unittest/*
	rm -rf \
	  bin \
	  $(BUILD_DIR)

shellspec:
	go run cmd/shellspec/main.go

validate-hardware-type-schemas:
	go run ./pkg/hardwaretypes/validate pkg/hardwaretypes/hardware-types/schema  pkg/hardwaretypes/hardware-types/

sim-setup:
	@if ! [ -d hms-simulation-environment ]; then \
	git clone https://github.com/Cray-HPE/hms-simulation-environment.git; \
	spec/support/bin/setup_simulator.sh ./hms-simulation-environment ../testdata/fixtures/sls/no_hardware.json; \
	fi

load-sls:
	spec/support/bin/setup_simulator.sh ./hms-simulation-environment ../testdata/fixtures/sls/no_hardware.json

spec-setup:
	@if ! [ -d ./shellspec ]; then \
	git clone https://github.com/shellspec/shellspec.git; \
	ln -s "$(shell pwd)"/shellspec/shellspec /usr/local/bin/; \
	fi

unittest: bin
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -cover \
	     github.com/Cray-HPE/cani/internal/inventory \
	     github.com/Cray-HPE/cani/internal/provider/csm \
	     github.com/Cray-HPE/cani/internal/provider/csm/ipam \
	     github.com/Cray-HPE/cani/internal/provider/csm/sls \
	     github.com/Cray-HPE/cani/internal/provider/csm/validate \
	     github.com/Cray-HPE/cani/internal/provider/csm/validate/checks \
	     github.com/Cray-HPE/cani/internal/provider/csm/validate/common

functional: bin
	./spec/support/bin/cani_integrate.sh functional

integrate: bin
	./spec/support/bin/cani_integrate.sh integration

edge: bin
	./spec/support/bin/cani_integrate.sh edge

test: bin validate-hardware-type-schemas unittest functional integrate edge

tools:
	go install golang.org/x/lint/golint@latest
	go install github.com/t-yuki/gocover-cobertura@latest
	go install github.com/jstemmer/go-junit-report@latest
	go install golang.org/x/tools/cmd/goimports@latest

bin/swagger-codegen-cli.jar:
	mkdir -p ./bin
	wget https://repo1.maven.org/maven2/io/swagger/codegen/v3/swagger-codegen-cli/3.0.43/swagger-codegen-cli-3.0.43.jar -O bin/swagger-codegen-cli.jar

# needs human munging until go-jsonschema can read a dir for resolving refs
netbox-dt-schema:
	go-jsonschema -p netbox devicetype-library/schema/devicetype.json -o pkg/netbox/types_devicetypes.go

# needs human munging until go-jsonschema can read a dir for resolving refs
netbox-mt-schema:
	go-jsonschema -p netbox devicetype-library/schema/moduletype.json -o pkg/netbox/types_moduletypes.go

# needs human munging find/replace to make the generated file work
nbschema: netbox-dt-schema netbox-mt-schema
	go-jsonschema -p netbox pkg/netbox/schema/devicetype-for-go-jsonschema.json -o pkg/netbox/types_devicetypes.go
	go-jsonschema -p netbox pkg/netbox/schema/moduletype-for-go-jsonschema.json -o pkg/netbox/types_moduletypes.go

vet: version
	go vet -v ./...

lint: tools
	golint -set_exit_status ./cmd/...
	golint -set_exit_status ./internal/...
	golint -set_exit_status ./pkg/...

fmt:
	go fmt ./...

env:
	@go env

tidy:
	go mod tidy

generate-go:
	go generate ./...

# Generate clients from the following swagger files:
# System Layout Service: ./pkg/sls-client/openapi.yaml
generate-swagger-sls-client: bin/swagger-codegen-cli.jar
	java -jar bin/swagger-codegen-cli.jar generate -i ./pkg/sls-client/openapi.yaml -l go -o ./pkg/sls-client/ -DpackageName=sls_client -t ./pkg/sls-client/templates
	go fmt ./pkg/sls-client/...
	goimports -w ./pkg/sls-client

# Generate clients from the following swagger files:
# Hardware State Manager: ./pkg/hsm-client/openapi.yaml
generate-swagger-hsm-client: bin/swagger-codegen-cli.jar
	java -jar bin/swagger-codegen-cli.jar generate -i ./pkg/hsm-client/openapi.yaml -l go -o ./pkg/hsm-client/ -DpackageName=hsm_client
	go fmt ./pkg/hsm-client/...
	goimports -w ./pkg/hsm-client

# Generate clients from the following swagger files:
# HPCM: ./pkg/hpcm-client/openapi.yaml
generate-swagger-hpcm-client: bin/swagger-codegen-cli.jar
	java -jar bin/swagger-codegen-cli.jar generate -i ./pkg/hpcm-client/openapi.yml -l go -o ./pkg/hpcm-client/ -DpackageName=hpcm_client
	go fmt ./pkg/hpcm-client/...
	goimports -w ./pkg/hpcm-client
	
venv:
	virtualenv -p python3 venv
	source venv/bin/activate
	pip install -r requirements.txt

generate-hardwaretypes-docs-mac:
	mkdir -p docs/hardware-types
	generate-schema-doc --config-file docs/generate-schema-doc-config.yml pkg/hardwaretypes/hardware-types/schema/devicetype.json docs/hardware-types/devicetype.md
	sed -i '' 's/Must be one of:/Must be one of:\n/g' docs/hardware-types/devicetype.md

# this has a stupid hack to make the markdown lists show properly by sed'ing the 
# resultant file.  this hack also needs a short delay for the generated file to 
# fully appear so it can be hacked
generate-hardwaretypes-docs:
	mkdir -p docs/hardware-types
	generate-schema-doc --config-file docs/generate-schema-doc-config.yml pkg/hardwaretypes/hardware-types/schema/devicetype.json docs/hardware-types/devicetype.md
	sed -i 's/Must be one of:/Must be one of:\n/g' docs/hardware-types/devicetype.md

generate: generate-swagger-sls-client generate-swagger-hsm-client generate-go generate-hardwaretypes-docs

serve: venv
	mkdocs serve

license:
	docker run -it --rm -v ${PWD}:/github/workspace artifactory.algol60.net/csm-docker/stable/license-checker .github/workflows/ cmd/ internal pkg/hardwaretypes pkg/xname spec/ --fix

# Jenkins doesn't have java installed, so the generate target fails to run
bin:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/${NAME} -ldflags '$(go_ldflags)'

rpm_prepare:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/SPECS $(BUILD_DIR)/SOURCES
	cp $(SPEC_FILE) $(BUILD_DIR)/SPECS/

rpm_package_source:
	tar --transform 'flags=r;s,^,/$(SOURCE_NAME)/,' --exclude .git --exclude dist -cvjf $(SOURCE_PATH) .

rpm_build_source:
	rpmbuild --nodeps --target $(ARCH) -ts $(SOURCE_PATH) --define "_topdir $(BUILD_DIR)"

rpm_build:
	rpmbuild --nodeps --target $(ARCH) -ba $(SPEC_FILE) --define "_topdir $(BUILD_DIR)"

doc:
	godoc -http=:8080 -index

version:
	@go version
