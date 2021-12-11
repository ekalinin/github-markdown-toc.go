EXEC=gh-md-toc
CMD_SRC=cmd/${EXEC}/main.go
BUILD_DIR=build
BUILD_OS="windows darwin linux"
BUILD_ARCH="amd64"

clean:
	@rm -f ${EXEC}
	@rm -f ${BUILD_DIR}/*
	@go clean

lint:
	@golint
	@golangci-lint run

# make run ARGS="--help"
run:
	@go run ${CMD_SRC} $(ARGS)

build: clean lint
	go build -race -o ${EXEC} ${CMD_SRC}

test: clean lint
	@go test -cover -o ${EXEC}

release: test buildall
	@git tag `grep "version" main.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`
	@git push --tags origin master

buildall: clean
	@mkdir -p ${BUILD_DIR}
	@for os in "${BUILD_OS}" ; do \
		for arch in "${BUILD_ARCH}" ; do \
			echo " * build $$os for $$arch"; \
			GOOS=$$os GOARCH=$$arch go build -o ${BUILD_DIR}/${EXEC} ${CMD_SRC}; \
			cd ${BUILD_DIR}; \
			tar czf ${EXEC}.$$os.$$arch.tgz ${EXEC}; \
			cd - ; \
		done done
	@rm ${BUILD_DIR}/${EXEC}
