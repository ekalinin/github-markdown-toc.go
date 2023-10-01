EXEC=gh-md-toc
CMD_SRC=cmd/${EXEC}/main.go
BUILD_DIR=build
BUILD_OS="windows darwin linux"
BUILD_ARCH="amd64"
E2E_DIR=e2e-tests

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
	go run cmd/gh-md-toc/main.go ./README.md > ${E2E_DIR}/got.md
	diff ${E2E_DIR}/want.md ${E2E_DIR}/got.md

release: test
	@git tag v`grep "\tVersion" internals.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`
	@git push --tags origin master

release-local:
	@goreleaser check
	@goreleaser release --snapshot --clean
