EXEC=gh-md-toc
CMD_SRC=cmd/${EXEC}/main.go
BUILD_DIR=build
E2E_DIR=e2e-tests
E2E_RUN=go run ./cmd/${EXEC} ./README.md
E2E_RUN_RHTML=go run ./cmd/${EXEC} https://github.com/ekalinin/github-markdown-toc.go/blob/master/README.md
E2E_RUN_RMD=go run ./cmd/${EXEC} https://raw.githubusercontent.com/ekalinin/github-markdown-toc.go/master/README.md
VERSION=$(shell grep "\tVersion" internal/version/version.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}')
TAG=v${VERSION}
bold := $(shell tput bold)
clear := $(shell tput sgr0)

clean:
	@rm -f ${EXEC}
	@rm -f ${BUILD_DIR}/*
	@go clean

lint:
	@go vet ./...
	@golangci-lint run ./...

# make run ARGS="--help"
run:
	@go run ${CMD_SRC} $(ARGS)

build: clean lint
	go build -race -o ${EXEC} ${CMD_SRC}

test: clean lint
	@go test -cover ./...

test-cover:
	@go test -covermode=count -coverprofile .cover.out ./internal/... | sort
	@go tool cover -html .cover.out -o .coverage.html

e2e:
	@echo "${bold}>> 1. Local MD, with options ...${clear}"
	${E2E_RUN} > ${E2E_DIR}/got.md
	@diff ${E2E_DIR}/want.md ${E2E_DIR}/got.md
	${E2E_RUN} --hide-header --hide-footer --depth=1 --no-escape > ${E2E_DIR}/got2.md
	@diff ${E2E_DIR}/want2.md ${E2E_DIR}/got2.md
	${E2E_RUN} --hide-header --hide-footer --indent=4 > ${E2E_DIR}/got3.md
	@diff ${E2E_DIR}/want3.md ${E2E_DIR}/got3.md

	@echo "${bold}>> 2. Remote MD, with options ...${clear}"
	${E2E_RUN_RMD} > ${E2E_DIR}/got4.md
	@diff ${E2E_DIR}/want.md ${E2E_DIR}/got4.md
	${E2E_RUN_RMD} --hide-header --hide-footer --depth=1 --no-escape > ${E2E_DIR}/got5.md
	@diff ${E2E_DIR}/want2.md ${E2E_DIR}/got5.md
	${E2E_RUN_RMD} --hide-header --hide-footer --indent=4 > ${E2E_DIR}/got6.md
	@diff ${E2E_DIR}/want3.md ${E2E_DIR}/got6.md

	@echo "${bold}>> 3. Remote HTML, with options ...${clear}"
	${E2E_RUN_RHTML} > ${E2E_DIR}/got7.md
	@diff ${E2E_DIR}/want.md ${E2E_DIR}/got7.md
	${E2E_RUN_RHTML} --hide-header --hide-footer --depth=1 --no-escape > ${E2E_DIR}/got8.md
	@diff ${E2E_DIR}/want2.md ${E2E_DIR}/got8.md
	${E2E_RUN_RHTML} --hide-header --hide-footer --indent=4 > ${E2E_DIR}/got9.md
	@diff ${E2E_DIR}/want3.md ${E2E_DIR}/got9.md

	@echo "${bold}>> 4. Multiple files, links carry the document path ...${clear}"
	go run ./cmd/${EXEC} --hide-footer ./README.md ./CHANGELOG.md > ${E2E_DIR}/got-combo.md
	@grep -qF '](./README.md#' ${E2E_DIR}/got-combo.md
	@grep -qF '](./CHANGELOG.md#' ${E2E_DIR}/got-combo.md

# Step 2: create the release tag locally. Does not push anything.
release: test release-local
	@if git rev-parse -q --verify refs/tags/${TAG} >/dev/null; then \
		echo "${bold}tag ${TAG} already exists${clear}"; \
		exit 1; \
	fi
	@git tag ${TAG}
	@echo "${bold}>> tag ${TAG} created, run 'make release-push' to publish${clear}"

# Step 1: validate the release locally. Creates neither tag nor push.
release-local:
	@goreleaser check
	@goreleaser release --snapshot --clean

# Step 3: publish the tag, which triggers the goreleaser workflow.
release-push:
	@git push origin ${TAG}
