GOPATH:=`pwd`/vendor:$(GOPATH)
GOPATH:=$(GOPATH):`pwd`/vendor/src/github.com/opencontainers/runc/libcontainer/
GOPATH:=$(GOPATH):`pwd`/vendor/src/github.com/opencontainers/runc/Godeps/_workspace
GO:=$(shell which go)
VERSION:=1.0
HARDWARE=$(shell uname -m)

build: vendor
	GOPATH=$(GOPATH) go build -ldflags="-X main.version $(VERSION)"
	GOPATH=$(GOPATH) bash -c 'cd psdock-ls && go build'

integration-test:
	GOPATH=$(GOPATH) go build -ldflags="-X main.version $(VERSION)"
	sudo PATH=$(PATH):`pwd` GOPATH=$(GOPATH) $(GO) test

test:
	GOPATH=$(GOPATH) bash -c 'cd logrotate && go test -cover'
	GOPATH=$(GOPATH) bash -c 'cd stream && go test -cover'
	sudo GOPATH=$(GOPATH) bash -c 'cd fsdriver && $(GO) test -cover'
	sudo PATH=$(PATH):`pwd` GOPATH=$(GOPATH) bash -c 'cd system && $(GO) test -cover'
	sudo GO_ENV=testing PATH=$(PATH):`pwd` GOPATH=$(GOPATH) bash -c 'cd integration && $(GO) test'

release:
	#psdock
	mkdir -p release
	GOPATH=$(GOPATH) GOOS=linux go build -ldflags="-X main.version $(VERSION)" -o release/psdock
	cd release && tar -zcf psdock-v$(VERSION)_$(HARDWARE).tgz psdock
	rm release/psdock

	#psdock-ls
	GOPATH=$(GOPATH) GOOS=linux bash -c 'cd psdock-ls && go build -o ../release/psdock-ls'
	cd release && tar -zcf psdock-ls-v$(VERSION)_$(HARDWARE).tgz psdock-ls
	rm release/psdock-ls

clean:
	rm -rf ./psdock ./release ./vendor/pkg

vendor:
	mkdir -p ./vendor/src/github.com/applidget
	ln -s `pwd` ./vendor/src/github.com/applidget/
	sh vendor.sh
