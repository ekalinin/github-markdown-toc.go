EXEC=gh-md-toc
BUILD_DIR=build
BUILD_OS="windows darwin freebsd linux"
BUILD_ARCH="amd64 386"

env:
	@nv mk github-toc --go-prebuilt=1.4.2 --force

env-activate:
	nv use github-toc

clean:
	@rm -f ${EXEC}
	@rm -f ${BUILD_DIR}/*
	@go clean

get-deps:
	@go get gopkg.in/alecthomas/kingpin.v2

build: clean
	@go build -ldflags "-s" -o ${EXEC}

test: clean
	@go test -cover -o ${EXEC}

release: buildall
	@git tag `grep "version" main.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`
	@git push --tags origin master

toolchain:
	@cd `go env GOROOT`/src 	&& \
		GOOS=windows GOARCH=amd64 CGO_ENABLED=0 ./make.bash --no-clean 	&& \
		GOOS=windows GOARCH=386   CGO_ENABLED=0 ./make.bash --no-clean 	&& \
		GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 ./make.bash --no-clean 	&& \
		GOOS=darwin  GOARCH=386   CGO_ENABLED=0 ./make.bash --no-clean 	&& \
		GOOS=linux   GOARCH=386   CGO_ENABLED=0 ./make.bash --no-clean 	&& \
		GOOS=freebsd GOARCH=386   CGO_ENABLED=0 ./make.bash --no-clean 	&& \
		GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 ./make.bash --no-clean

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
