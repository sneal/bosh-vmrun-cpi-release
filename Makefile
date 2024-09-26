NAME ?= bosh-vmrun-cpi
RELEASE_DIR = ./dist
OUTPUT = ./dist/cpi.tgz
VERSION ?= 0.0.0
GITSHA = $(shell git rev-parse HEAD)
GITDIRTY = $(shell git diff --quiet HEAD || echo "dirty")

.PHONY: all
all: clean release

.PHONY: clean
clean:
	@rm -rf dist/

.PHONY: release
release: $(GO_SOURCES) ## Cross-compile binary for various operating systems
	@mkdir -p $RELEASE_DIR
	bosh create-release --sha2 --force --dir . --tarball $OUTPUT

