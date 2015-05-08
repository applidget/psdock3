GOPATH:=`pwd`/vendor:$(GOPATH)
GOPATH:=$(GOPATH):`pwd`/vendor/src/github.com/docker/libcontainer/vendor
GO:=$(shell which go)
VERSION:=0.1
HARDWARE=$(shell uname -m)

build: vendor
	GOPATH=$(GOPATH) go build

integration-test:
	GOPATH=$(GOPATH) go build
	sudo PATH=$(PATH):`pwd` GOPATH=$(GOPATH) $(GO) test

test:
	GOPATH=$(GOPATH) bash -c 'cd logrotate && go test'
	GOPATH=$(GOPATH) bash -c 'cd portwatcher && go test'
	GOPATH=$(GOPATH) bash -c 'cd stream && go test'
	sudo GOPATH=$(GOPATH) bash -c 'cd fsdriver && $(GO) test'
	sudo PATH=$(PATH):`pwd` GOPATH=$(GOPATH) bash -c 'cd proc && $(GO) test'
	sudo GO_ENV=testing PATH=$(PATH):`pwd` GOPATH=$(GOPATH) bash -c 'cd integration && $(GO) test'

release:
	mkdir -p release
	GOPATH=$(GOPATH) GOOS=linux go build -o release/psdock
	cd release && tar -zcf psdock-v$(VERSION)_$(HARDWARE).tgz psdock
	rm release/psdock

clean:
	rm -rf ./psdock ./release ./vendor/pkg

vendor:
	mkdir -p ./vendor/src/github.com/robinmonjo
	ln -s `pwd` ./vendor/src/github.com/robinmonjo/
	sh vendor.sh
