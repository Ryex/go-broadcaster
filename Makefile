
#TOOLS

GO			=	go
GODEP		=	dep
GOEXE 	:= $(shell go env GOEXE)

NPM			= yarn
NPMRUN	= yarn

MODE ?= production

default: build

build: mediamonitor web

mediamonitor: _godeps
	$(GO) build -o ./bin/media-monitor$(GOEXE) ./cmd/media-monitor

web: webserver

webserver: _godeps webclient
ifneq ($(MODE), production)
	$(GO) build -o ./bin/webserver$(GOEXE) -tags=dev ./web
else
	$(GO)	generate ./web/client
	$(GO) build -o ./bin/webserver$(GOEXE) ./web
endif

webclient: _npmdeps
ifneq ($(MODE), production)
	cd ./web/client && $(NPMRUN) dev-build
else
	cd ./web/client && $(NPMRUN) build
endif
	cp -r ./web/client/dist ./bin

_godeps: go.mod
	$(GO) mod download

_npmdeps:
	$(NPM) install

clean:
	rm -rf ./bin
