EXEC=gh-md-toc
BUILD_DIR=build
BUILD_OS="windows darwin freebsd linux"
BUILD_ARCH="amd64 386"
GOVER=1.7.3
ENVNAME=github-toc${GOVER}

env-init:
	@bash -c ". ~/.envirius/nv && nv mk ${ENVNAME} --go-prebuilt=${GOVER}"

# with https://github.com/ekalinin/envirius
env:
	@bash -c ". ~/.envirius/nv && nv use ${ENVNAME}"

clean:
	@rm -f ${EXEC}
	@rm -f ${BUILD_DIR}/*
	@go clean

get-deps:
	@go get gopkg.in/alecthomas/kingpin.v2

# http://tschottdorf.github.io/linking-golang-go-statically-cgo-testing/
build: clean
	@go build -a -tags netgo --ldflags '-s -extldflags "-lm -lstdc++ -static"' -i -o ${EXEC}

test: clean
	@go test -cover -o ${EXEC}

release: test buildall
	@git tag `grep "version" main.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`
	@git push --tags origin master

buildall: clean
	@mkdir -p ${BUILD_DIR}
	@for os in "${BUILD_OS}" ; do \
		for arch in "${BUILD_ARCH}" ; do \
			echo " * build $$os for $$arch"; \
			GOOS=$$os GOARCH=$$arch go build -ldflags "-s" -o ${BUILD_DIR}/${EXEC}; \
			cd ${BUILD_DIR}; \
			tar czf ${EXEC}.$$os.$$arch.tgz ${EXEC}; \
			cd - ; \
		done done
	@rm ${BUILD_DIR}/${EXEC}
