EXEC=gh-md-toc
CMD_SRC=cmd/${EXEC}/main.go
BUILD_DIR=build
BUILD_OS="windows darwin linux"
BUILD_ARCH="amd64"
E2E_DIR=e2e-tests
E2E_RUN=go run cmd/gh-md-toc/main.go ./README.md
E2E_RUN_RHTML=go run cmd/gh-md-toc/main.go https://github.com/ekalinin/github-markdown-toc.go/blob/master/README.md
E2E_RUN_RMD=go run cmd/gh-md-toc/main.go https://raw.githubusercontent.com/ekalinin/github-markdown-toc.go/master/README.md
bold := $(shell tput bold)
clear := $(shell tput sgr0)

clean:
	@rm -f ${EXEC}
	@rm -f ${BUILD_DIR}/*
	@go clean

lint:
	@go vet
	@golangci-lint run

# make run ARGS="--help"
run:
	@go run ${CMD_SRC} $(ARGS)

build: clean lint
	go build -race -o ${EXEC} ${CMD_SRC}

test: clean lint
	@go test -cover -o ${EXEC}

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

release: test
	@git tag v`grep "\tVersion" internal/version.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`
	@git push --tags origin master

release-local:
	@goreleaser check
	@goreleaser release --snapshot --clean
