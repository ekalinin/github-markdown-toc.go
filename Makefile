EXEC=gh-md-toc

env:
	@nv mk github-toc --go-prebuilt=1.4.2 --force

env-activate:
	nv use github-toc

clean:
	@rm -f ${EXEC}*
	@go clean

build: clean
	@go build -o ${EXEC}

buildall: clean
	GOARCH=amd64 go build -o ${EXEC}_amd64
	GOARCH=386 go build -o ${EXEC}_i386

buildstripped: clean
	@go build -ldflags "-s" -o ${EXEC}

test: clean
	@go test -cover fmt -o ${EXEC}

release:
	@git tag `grep "version =" main.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`
	@#git push --tags origin master
