GOPATH:=`pwd`/vendor:$(GOPATH)
GOPATH:=$(GOPATH):`pwd`/vendor/src/github.com/docker/libcontainer/vendor
GO:=$(shell which go)
VERSION:=0.1
HARDWARE=$(shell uname -m)

build: vendor
	GOPATH=$(GOPATH) go build

test:
	GOPATH=$(GOPATH) go build
	sudo PATH=$(PATH):`pwd` GOPATH=$(GOPATH) $(GO) test

release:
	mkdir -p release
	GOPATH=$(GOPATH) GOOS=linux go build -o release/psdock
	cd release && tar -zcf psdock-v$(VERSION)_$(HARDWARE).tgz psdock
	rm release/psdock

clean:
	rm -rf ./psdock ./release ./vendor/pkg

vendor:
	sh vendor.sh
